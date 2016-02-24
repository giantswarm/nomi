package output

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aybabtme/uniplot/histogram"

	"github.com/giantswarm/fleemmer/log"
	"github.com/giantswarm/fleemmer/unit"
)

// DumpJSON prints the metrics of the benchmark in a JSON format
func DumpJSON(stats unit.Stats) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(stats)
}

// DumpJSON dumps the stats metrics to a javascript file 'data.js' which should
// be used by embedded scripts to print a graphic.
func DumpHTMLTar(html []byte, scriptJs []byte, stats unit.Stats) {
	jsonData := bytes.NewBufferString("var allData = ")
	enc := json.NewEncoder(jsonData)
	enc.Encode(stats)
	jsonData.WriteString(";\n")

	tw := tar.NewWriter(os.Stdout)
	var files = []struct {
		Name string
		Body []byte
	}{
		{"data.js", jsonData.Bytes()},
		{"index.html", html},
		{"script.js", scriptJs},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name:       file.Name,
			ChangeTime: time.Now(),
			ModTime:    time.Now(),
			Mode:       0644,
			Size:       int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Logger().Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Logger().Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		log.Logger().Fatal(err)
	}

}

// PrintReport prints in stdout a report of the the units delay for the start operation.
func PrintReport(stats unit.Stats, out io.Writer) {
	delays := []float64{}
	maxRunningCount := 0
	minStartingTime := 10000000.0
	minDelay := 10000000.0
	maxDelay := 0.0
	maxCompletionTime := 0.0

	for _, ev := range stats.Start {
		if maxRunningCount <= ev.RunningCount {
			maxRunningCount = ev.RunningCount
		}
		if minDelay >= ev.Delay {
			minDelay = ev.Delay
		}
		if maxDelay <= ev.Delay {
			maxDelay = ev.Delay
		}
		if minStartingTime > ev.StartTime {
			minStartingTime = ev.StartTime
		}
		if maxCompletionTime <= ev.CompletionTime {
			maxCompletionTime = ev.CompletionTime
		}

		delays = append(delays, ev.Delay)
	}
	hist := histogram.Hist(10, delays)

	fmt.Println("Number of runnings units: ", maxRunningCount)
	fmt.Println("Minimum time to start an unit: ", minDelay)
	fmt.Println("Maximum time to start an unit: ", maxDelay)
	fmt.Println("Time to compute the start operation (secs): ", float64(maxCompletionTime-minStartingTime))

	fmt.Println("-- Histogram Starting Delay --")
	histogram.Fprint(out, hist, histogram.Linear(20))
}
