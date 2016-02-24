package unit

import (
	"log"
	"testing"

	"github.com/giantswarm/fleemmer/definition"
)

func TestEngine(t *testing.T) {
	rawInstructions := "(sleep 1) (start 2700 100) (sleep 700) (stop-all)"
	def, err := definition.BenchmarkDefByRawInstructions(rawInstructions, 3)
	engine, err2 := NewEngine(def, false)

	if err2 != nil {
		log.Fatalf("unable to create the new engine: %v", err)
	}

	if engine.InstanceGroupSize() != 3 {
		log.Fatalf("wrong instance group size expected '3' got: %s", engine.InstanceGroupSize())
	}
}
