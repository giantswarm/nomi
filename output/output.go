package output

import (
	"archive/tar"
	"bytes"
	"encoding/json"
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

// PrintHistogram prints in stdout the units delay for the start operation
func PrintHistogram(stats unit.Stats, out io.Writer) {
	delays := []float64{}
	for _, ev := range stats.Start {
		delays = append(delays, ev.Delay)
	}
	hist := histogram.Hist(10, delays)
	histogram.Fprint(out, hist, histogram.Linear(20))
}
