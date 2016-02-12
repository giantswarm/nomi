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
	"github.com/golang/glog"

	"github.com/giantswarm/fleemmer/unit"
)

func DumpJSON(stats unit.Stats) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(stats)
}

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
			glog.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			glog.Fatalln(err)
		}
	}

	if err := tw.Close(); err != nil {
		glog.Fatalln(err)
	}

}

// Print the units delay to start
func PrintHistogram(stats unit.Stats, out io.Writer) {
	delays := []float64{}
	for _, ev := range stats.Start {
		delays = append(delays, ev.Delay)
	}
	hist := histogram.Hist(10, delays)
	fmt.Println(">> Histogram Starting Units Delay <<")
	histogram.Fprint(out, hist, histogram.Linear(20))
}
