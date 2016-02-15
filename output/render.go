// This output package contains all the operations to render the metrics collected
// during the benchmark
package output

import (
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"

	"github.com/giantswarm/fleemmer/output/gnuplot"
	"github.com/giantswarm/fleemmer/unit"
)

// Directory where the plots are stored by default
const plotsDIR = "/fleemmer_plots"

var processTypes = []string{"fleetd", "systemd"}

// GeneratePlots creates some initial plots from the collected metrics. Three
// are the initial plots: start operation completion time/delay, stop operation
// completion time/delay and cluster metrics for systemd and fleetd.
func GeneratePlots(stats unit.Stats) {
	fname := ""
	persist := true
	debug := false
	plotsDirectory := plotsDIR

	if os.Getenv("PLOTS_DIR") != "" {
		plotsDirectory = os.Getenv("PLOTS_DIR")
	}

	gnuplot.Initialize()

	generateDelayStartPlot(fname, persist, debug, plotsDirectory, stats)

	// Start delay
	if len(stats.Start) > 0 {
		generateUnitsStopPlot(fname, persist, debug, plotsDirectory, stats)
	}

	// Stop delay
	if len(stats.Stop) > 0 {
		generateUnitsStopPlot(fname, persist, debug, plotsDirectory, stats)
	}
}

func generateDelayStartPlot(fname string, persist bool, debug bool, plotsDirectory string, stats unit.Stats) {
	for _, process := range processTypes {
		p, err := gnuplot.NewPlotter(fname, persist, debug)
		if err != nil {
			err_string := fmt.Sprintf("** err: %v\n", err)
			panic(err_string)
		}
		defer p.Close()
		for hostname, metrics := range stats.MachineStats {
			valuesX := make([]float64, 0)
			valuesY := make([]float64, 0)
			for _, metric := range metrics {
				if metric.Process == process {
					valuesX = append(valuesX, metric.TimeStamp)
					valuesY = append(valuesY, metric.CPUUsage)
				}
			}
			glog.V(2).Infof("Plotting date for %s", hostname)
			p.PlotXY(valuesX, valuesY, fmt.Sprintf("%s - Time/CPU", hostname))
		}
		p.SetXLabel("Timestamp")
		p.SetYLabel("CPU usage")
		p.CheckedCmd("set terminal pdf")

		glog.V(2).Infof("Generating plot %s", process)
		p.CheckedCmd(fmt.Sprintf("set output '%s/%s.pdf'", plotsDirectory, process))
		p.CheckedCmd("replot")

		time.Sleep(2)
		p.CheckedCmd("q")

	}
}

func generateUnitsStartPlot(fname string, persist bool, debug bool, plotsDirectory string, stats unit.Stats) {
	p, err := gnuplot.NewPlotter(fname, persist, debug)
	if err != nil {
		err_string := fmt.Sprintf("** err: %v\n", err)
		panic(err_string)
	}
	defer p.Close()

	valuesX := make([]float64, 0)
	valuesY := make([]float64, 0)
	p.CheckedCmd("set grid x")
	p.CheckedCmd("set grid y")
	p.SetStyle("impulses")
	for _, stats := range stats.Start {
		valuesX = append(valuesX, stats.CompletionTime)
		valuesY = append(valuesY, stats.Delay)
	}
	p.PlotXY(valuesX, valuesY, "Stop operation Completion/Delay seconds")
	p.SetXLabel("Completion time")
	p.SetYLabel("Delay time")
	p.CheckedCmd("set terminal pdf")

	p.CheckedCmd(fmt.Sprintf("set output '%s/units_start.pdf'", plotsDirectory))
	p.CheckedCmd("replot")

	time.Sleep(2)
	p.CheckedCmd("q")
}

func generateUnitsStopPlot(fname string, persist bool, debug bool, plotsDirectory string, stats unit.Stats) {
	p, err := gnuplot.NewPlotter(fname, persist, debug)
	if err != nil {
		err_string := fmt.Sprintf("** err: %v\n", err)
		panic(err_string)
	}
	defer p.Close()

	valuesX := make([]float64, 0)
	valuesY := make([]float64, 0)
	p.CheckedCmd("set grid x")
	p.CheckedCmd("set grid y")
	p.SetStyle("impulses")
	p.SetYLabel("Delay time")
	p.SetXLabel("Completion time")
	p.CheckedCmd("set terminal pdf")
	p.CheckedCmd(fmt.Sprintf("set output '%s/units_stop.pdf'", plotsDirectory))

	for _, stats := range stats.Stop {
		valuesX = append(valuesX, stats.CompletionTime)
		valuesY = append(valuesY, stats.Delay)
	}
	p.PlotXY(valuesX, valuesY, "Stop operation Completion/Delay seconds")
	p.CheckedCmd("replot")

	time.Sleep(2)
	p.CheckedCmd("q")
}
