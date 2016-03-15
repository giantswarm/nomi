package cmd

import (
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/giantswarm/nomi/definition"
	"github.com/giantswarm/nomi/fleet"
	"github.com/giantswarm/nomi/log"
	"github.com/giantswarm/nomi/output"
	"github.com/giantswarm/nomi/unit"
)

const (
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
	verbose         bool
	unitFile        string
}

func (f runCmdFlags) Validate() {
	if f.dumpJSONFlag && f.dumpHTMLTarFlag {
		log.Logger().Fatal("dump option is required. Please, choose between:  dump-json OR dump-html-tar")
	}

	if f.generatePlots {
		if _, err := exec.LookPath("gnuplot"); err != nil {
			log.Logger().Infof("generate-gnuplots: could not find path to 'gnuplot':\n%v\n", err)
			log.Logger().Fatal("generate-gnuplots option requires 'gnuplot' software installed")
		}
	}

	if f.benchmarkFile == "" && f.rawInstructions == "" {
		log.Logger().Fatal("benchmark file definition or raw instructions is required")
	}

	if f.benchmarkFile != "" && f.rawInstructions != "" {
		log.Logger().Fatal("benchmark file definition or raw instructions are mutual exclusive")
	}

	if f.benchmarkFile == "" && f.rawInstructions != "" && f.igSize <= 0 {
		log.Logger().Fatal("instance group size has to be greater than 0 when using raw-instructions parameter")
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
	runCmd.Flags().StringVar(&runFlags.benchmarkFile, "benchmark-file", "", "file with the benchmark definition (application definition, instance group size, instructions to spawn/stop/float units)")
	runCmd.Flags().StringVar(&runFlags.rawInstructions, "raw-instructions", "", "instructions to spawn/stop/float units")
	runCmd.Flags().BoolVar(&runFlags.dumpJSONFlag, "dump-json", false, "dump json stats to stdout")
	runCmd.Flags().BoolVar(&runFlags.dumpHTMLTarFlag, "dump-html-tar", false, "dump tarred html stats to stdout")
	runCmd.Flags().BoolVar(&runFlags.verbose, "verbose", false, "verbose output")
	runCmd.Flags().BoolVar(&runFlags.generatePlots, "generate-gnuplots", false, "generate plots using gnuplot (output directory=/nomi_plots)")
	runCmd.Flags().IntVar(&runFlags.igSize, "instancegroup-size", 1, "instance group size")
}

func runRun(cmd *cobra.Command, args []string) {
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
		log.Logger().Fatal(err)
	}

	var benchmark definition.BenchmarkDef
	if runFlags.benchmarkFile == "" {
		benchmark, err = definition.BenchmarkDefByRawInstructions(runFlags.rawInstructions, runFlags.igSize)
		if err != nil {
			log.Logger().Fatal("unable to parse the introduced raw instructions")
		}
	} else {
		benchmark, err = definition.BenchmarkDefByFile(runFlags.benchmarkFile)
	}

	if err != nil {
		log.Logger().Fatal(err)
	}

	unitEngine, err := unit.NewEngine(benchmark, runFlags.verbose)
	if err != nil {
		log.Logger().Fatal(err)
	}

	builder, err := unit.NewBuilder(benchmark.Application, unitEngine.InstanceGroupSize(), runFlags.listenAddr)
	if err != nil {
		log.Logger().Fatal(err)
	}

	if benchmark.Application.Type == "unitfiles" {
		err = builder.UseCustomUnitFileService(benchmark.Application.UnitFilePath)
		if err != nil {
			log.Logger().Fatal("unable to parse unit file from your application definition")
		}
	}

	observer := unit.NewUnitObserver(unitEngine)
	observer.StartHTTPService(runFlags.listenAddr)

	fleetPool.StartUnit(builder.MakeStatsDumper("etcd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -C etcd 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "etcd"))
	fleetPool.StartUnit(builder.MakeStatsDumper("fleetd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -C fleetd 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "fleetd"))
	fleetPool.StartUnit(builder.MakeStatsDumper("systemd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -p 1 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "systemd"))

	unitEngine.SpawnFunc = func(id string) error {
		if runFlags.verbose {
			log.Logger().Infof("spawning unit with id %s\n", id)
		}
		return fleetPool.StartUnitGroup(builder.MakeUnitChain(id))
	}

	unitEngine.StopFunc = func(id string) error {
		if runFlags.verbose {
			log.Logger().Infof("stopping unit with id %s\n", id)
		}
		return fleetPool.Stop(builder.GetUnitPrefix() + "-0@" + id + ".service")
	}

	unitEngine.Run()

	existingUnits, err = fleetPool.ListUnits()
	if err != nil {
		log.Logger().Errorf("error listing units %v", err)
	}

	wg := new(sync.WaitGroup)
	for _, unit := range existingUnits {
		if strings.HasPrefix(unit.Name, builder.GetUnitPrefix()) {
			wg.Add(1)
			go func(unitName string) {
				if runFlags.verbose {
					log.Logger().Infof("destroying old unit: %s", unitName)
				}
				fleetPool.Destroy(unitName)
				wg.Done()
			}(unit.Name)
		}
	}
	wg.Wait()

	generateBenchmarkReport(runFlags.dumpJSONFlag, runFlags.dumpHTMLTarFlag, runFlags.generatePlots, unitEngine)
}

func generateBenchmarkReport(dumpJSONFlag, dumpHTMLTarFlag, generatePlots bool, unitEngine *unit.UnitEngine) {
	if dumpJSONFlag {
		output.DumpJSON(unitEngine.Stats())
	}

	if dumpHTMLTarFlag {
		html, err := output.Asset("output/embedded/render.html")
		if err != nil {
			log.Logger().Fatal(err)
		}

		scriptJs, err := output.Asset("output/embedded/script.js")
		if err != nil {
			log.Logger().Fatal(err)
		}

		output.DumpHTMLTar(html, scriptJs, unitEngine.Stats())
	}

	if generatePlots {
		output.GeneratePlots(unitEngine.Stats(), runFlags.verbose)
	}

	output.PrintReport(unitEngine.Stats(), os.Stderr)
}
