package output

import (
	"fmt"
	"time"

	"bitbucket.org/binet/go-gnuplot/pkg/gnuplot"
	"github.com/golang/glog"

	"github.com/giantswarm/fleemmer/unit"
)

const plotsDIR = "/fleemmer_plots"

var processTypes = []string{"fleetd", "systemd"}

func GeneratePlots(stats unit.Stats) {
	fname := ""
	persist := true
	debug := false

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
		p.CheckedCmd(fmt.Sprintf("set output '%s/%s.pdf'", plotsDIR, process))
		p.CheckedCmd("replot")

		time.Sleep(2)
		p.CheckedCmd("q")

	}

	// Start delay
	if len(stats.Start) > 0 {
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

		p.CheckedCmd(fmt.Sprintf("set output '%s/units_start.pdf'", plotsDIR))
		p.CheckedCmd("replot")

		time.Sleep(2)
		p.CheckedCmd("q")
	}

	// Stop delay
	if len(stats.Stop) > 0 {
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
		p.CheckedCmd(fmt.Sprintf("set output '%s/units_stop.pdf'", plotsDIR))

		for _, stats := range stats.Stop {
			valuesX = append(valuesX, stats.CompletionTime)
			valuesY = append(valuesY, stats.Delay)
		}
		p.PlotXY(valuesX, valuesY, "Stop operation Completion/Delay seconds")
		p.CheckedCmd("replot")

		time.Sleep(2)
		p.CheckedCmd("q")
	}
}
