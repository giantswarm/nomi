// This main package contains the main logic of operations for fleemmer
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/golang/glog"

	"github.com/giantswarm/fleemmer/definition"
	"github.com/giantswarm/fleemmer/fleet"
	"github.com/giantswarm/fleemmer/output"
	"github.com/giantswarm/fleemmer/unit"
)

var (
	listenAddr      = flag.String("addr", "127.0.0.1:40302", "address to listen")
	benchmarkFile   = flag.String("benchmark-file", "", "file with the benchmark definition (instance group size, instructions to spawn/stop/float units)")
	rawInstructions = flag.String("raw-instructions", "", "instructions to spawn/stop/float units")
	dumpJSONFlag    = flag.Bool("dump-json", false, "dump json stats to stdout")
	dumpHTMLTarFlag = flag.Bool("dump-html-tar", false, "dump tarred html stats to stdout")
	generatePlots   = flag.Bool("generate-gnuplots", false, "generate plots using GNUPLOT (output directory=/fleemmer_plots)")
	igSize          = flag.Int("instancegroup-size", 1, "instance group size")
)

const beaconUnitPrefix = "beaconX"

func checkArgsSanity() {
	if *listenAddr == "" {
		glog.Fatalln("a non-empty address is required")
	}

	if *dumpJSONFlag && *dumpHTMLTarFlag {
		glog.Fatalln("dump option is required. Please, choose between:  dump-json OR dump-html-tar")
	}

	if !*dumpJSONFlag && !*dumpHTMLTarFlag {
		glog.Warningln("output mode option is required")
	}

	if *benchmarkFile == "" && *rawInstructions == "" {
		glog.Fatalln("benchmark file definition or raw instructions is required")
	}

	if *benchmarkFile != "" && *rawInstructions != "" {
		glog.Fatalln("benchmark file definition or raw instructions are mutual exclusive")
	}

	if *benchmarkFile == "" && *rawInstructions != "" && *igSize <= 0 {
		glog.Fatalln("instance group size has to be greater than 0 when using raw-instructions parameter")
	}

	if *generatePlots {
		if _, err := exec.LookPath("gnuplot"); err != nil {
			glog.V(2).Infof("generate-gnuplots: could not find path to 'gnuplot':\n%v\n", err)
			glog.Fatalln("generate-gnuplots option requires 'gnuplot' software installed")
		}
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("fleemmer: ")
	flag.Set("logtostderr", "true")
	flag.Parse()

	checkArgsSanity()
	fleetPool := fleet.NewFleetPool(20)

	existingUnits, err := fleetPool.ListUnits()
	if err != nil {
		glog.Fatalln(err)
	}

	var benchmark definition.BenchmarkDef
	if *benchmarkFile == "" {
		benchmark, err = definition.BenchmarkDefByRawInstructions(*rawInstructions, *igSize)
		if err != nil {
			glog.Fatalln("unable to parse the introduced raw instructions")
		}
	} else {
		benchmark, err = definition.BenchmarkDefByFile(*benchmarkFile)
	}
	if err != nil {
		glog.Fatalln(err)
	}

	unitEngine, err := unit.NewEngine(benchmark)
	if err != nil {
		glog.Fatalln(err)
	}

	observer := unit.NewBeaconObserver(unitEngine)
	observer.StartHTTPService(*listenAddr)

	fleetPool.StartUnit(unit.MakeStatsDumper("fleetd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -C fleetd 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "fleetd", *listenAddr))
	fleetPool.StartUnit(unit.MakeStatsDumper("systemd", "echo `hostname` `docker run --rm --pid=host ragnarb/toolbox pidstat -h -r -u -p 1 10 1 | tail -n 1 | awk \\'{print $7 \" \" $12}\\'`", "systemd", *listenAddr))

	unitEngine.SpawnFunc = func(id string) error {
		glog.V(2).Infof("spawning unit with id %s\n", id)
		return fleetPool.StartUnitGroup(unit.MakeUnitChain(id, *listenAddr, unitEngine.InstanceGroupSize()))
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

	if *generatePlots {
		output.GeneratePlots(unitEngine.Stats())
	}

	if *dumpJSONFlag {
		output.DumpJSON(unitEngine.Stats())
	}

	if *dumpHTMLTarFlag {
		html, err := Asset("output/embedded/render.html")
		if err != nil {
			glog.Fatalln(err)
		}

		scriptJs, err := Asset("output/embedded/script.js")
		if err != nil {
			glog.Fatalln(err)
		}

		output.DumpHTMLTar(html, scriptJs, unitEngine.Stats())
	}

	output.PrintHistogram(unitEngine.Stats(), os.Stderr)
}
