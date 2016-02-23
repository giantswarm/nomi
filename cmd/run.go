package cmd

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/giantswarm/fleemmer/definition"
	"github.com/giantswarm/fleemmer/fleet"
	"github.com/giantswarm/fleemmer/output"
	"github.com/giantswarm/fleemmer/unit"
)

const (
	beaconUnitPrefix = "beaconX"

	listenerDefaultIP   = "127.0.0.1"
	listenerDefaultPort = "40302"
)

type runCmdFlags struct {
	listenAddr      string
	benchmarkFile   string
	rawInstructions string
	dumpJSONFlag    bool
	dumpHTMLTarFlag bool
	generatePlots   bool
	igSize          int
}

func (f runCmdFlags) Validate() {
	if f.dumpJSONFlag && f.dumpHTMLTarFlag {
		glog.Fatalln("dump option is required. Please, choose between:  dump-json OR dump-html-tar")
	}

	if !f.dumpJSONFlag && !f.dumpHTMLTarFlag {
		glog.Warningln("output mode option is required")
	}

	if f.benchmarkFile == "" && f.rawInstructions == "" {
		glog.Fatalln("benchmark file definition or raw instructions is required")
	}

	if f.benchmarkFile != "" && f.rawInstructions != "" {
		glog.Fatalln("benchmark file definition or raw instructions are mutual exclusive")
	}

	if f.benchmarkFile == "" && f.rawInstructions != "" && f.igSize <= 0 {
		glog.Fatalln("instance group size has to be greater than 0 when using raw-instructions parameter")
	}

	if f.generatePlots {
		if _, err := exec.LookPath("gnuplot"); err != nil {
			glog.V(2).Infof("generate-gnuplots: could not find path to 'gnuplot':\n%v\n", err)
			glog.Fatalln("generate-gnuplots option requires 'gnuplot' software installed")
		}
	}
}

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run benchmark",
		Long:  "Run a fleet benchmark based on the provided configuration",
		Run:   runRun,
	}

	runFlags = runCmdFlags{}
)

func init() {
	runCmd.Flags().StringVar(&runFlags.listenAddr, "addr", "", "address to listen")
	runCmd.Flags().StringVar(&runFlags.benchmarkFile, "benchmark-file", "", "file with the benchmark definition (instance group size, instructions to spawn/stop/float units)")
	runCmd.Flags().StringVar(&runFlags.rawInstructions, "raw-instructions", "", "instructions to spawn/stop/float units")
	runCmd.Flags().BoolVar(&runFlags.dumpJSONFlag, "dump-json", false, "dump json stats to stdout")
	runCmd.Flags().BoolVar(&runFlags.dumpHTMLTarFlag, "dump-html-tar", false, "dump tarred html stats to stdout")
	runCmd.Flags().BoolVar(&runFlags.generatePlots, "generate-gnuplots", false, "generate plots using GNUPLOT (output directory=/fleemmer_plots)")
	runCmd.Flags().IntVar(&runFlags.igSize, "instancegroup-size", 1, "instance group size")
}

func runRun(cmd *cobra.Command, args []string) {
	log.SetFlags(0)
	log.SetPrefix("fleemmer: ")
	flag.Set("logtostderr", "true")

	runFlags.Validate()
	fleetPool := fleet.NewFleetPool(20)
	
	if runFlags.listenAddr == "" {
		// We extract the public CoreOS ip of the host machine
		ip, err := fleet.CoreosHostPublicIP()
		if ip == "" || err != nil {
			runFlags.listenAddr = listenerDefaultIP + ":" + listenerDefaultPort
		} else {
			runFlags.listenAddr = ip + ":" + listenerDefaultPort
		}
	}

	existingUnits, err := fleetPool.ListUnits()
	if err != nil {
		glog.Fatalln(err)
	}

	var benchmark definition.BenchmarkDef
	if runFlags.benchmarkFile == "" {
		benchmark, err = definition.BenchmarkDefByRawInstructions(runFlags.rawInstructions, runFlags.igSize)
		if err != nil {
			glog.Fatalln("unable to parse the introduced raw instructions")
		}
	} else {
		benchmark, err = definition.BenchmarkDefByFile(runFlags.benchmarkFile)
	}
	if err != nil {
		glog.Fatalln(err)
	}

	unitEngine, err := unit.NewEngine(benchmark)
	if err != nil {
		glog.Fatalln(err)
	}

	observer := unit.NewBeaconObserver(unitEngine)
	observer.StartHTTPService(runFlags.listenAddr)

	fleetPool.StartUnit(unit.MakeStatsDumper("fleetd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -C fleetd 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "fleetd", runFlags.listenAddr))
	fleetPool.StartUnit(unit.MakeStatsDumper("systemd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -p 1 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "systemd", runFlags.listenAddr))

	unitEngine.SpawnFunc = func(id string) error {
		glog.V(2).Infof("spawning unit with id %s\n", id)
		return fleetPool.StartUnitGroup(unit.MakeUnitChain(id, runFlags.listenAddr, unitEngine.InstanceGroupSize()))
	}

	unitEngine.StopFunc = func(id string) error {
		glog.V(2).Infof("stopping unit with id %s\n", id)
		return fleetPool.Stop(beaconUnitPrefix + "-0@" + id + ".service")
	}

	unitEngine.Run()

	existingUnits, err = fleetPool.ListUnits()
	if err != nil {
		glog.Errorf("error listing units %v", err)
	}

	wg := new(sync.WaitGroup)
	for _, unit := range existingUnits {
		if strings.HasPrefix(unit.Name, beaconUnitPrefix) {
			wg.Add(1)
			go func(unitName string) {
				glog.V(2).Infoln("destroying old unit: " + unitName)
				fleetPool.Destroy(unitName)
				wg.Done()
			}(unit.Name)
		}
	}
	wg.Wait()

	if runFlags.dumpJSONFlag {
		output.DumpJSON(unitEngine.Stats())
	}

	if runFlags.dumpHTMLTarFlag {
		html, err := output.Asset("output/embedded/render.html")
		if err != nil {
			glog.Fatalln(err)
		}

		scriptJs, err := output.Asset("output/embedded/script.js")
		if err != nil {
			glog.Fatalln(err)
		}

		output.DumpHTMLTar(html, scriptJs, unitEngine.Stats())
	}

	if runFlags.generatePlots {
		output.GeneratePlots(unitEngine.Stats())
	}

	output.PrintHistogram(unitEngine.Stats(), os.Stderr)
}
