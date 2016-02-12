package definition

import (
	"log"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

var dataTest = `
instancegroup-size: 1
instructions:
  - start:
     max: 1
     interval: 200
  - float:
     rate: 1.0
     duration: 110
  - expect-running:
     symbol: <
     amount: 10
  - sleep: 100
  - stop: stop-all
`

func TestInstructionYAMLDefinition(t *testing.T) {
	ins := BenchmarkDef{}

	if err := yaml.Unmarshal([]byte(dataTest), &ins); err != nil {
		log.Fatalf("unable to parse the yaml test definition: %v", err)
	}

	if ins.InstanceGroupSize != 1 {
		log.Fatalf("wrong instance group size is wrong %d expected 1", ins.InstanceGroupSize)
	}
	if len(ins.Instructions) != 5 {
		log.Fatalf("instructions size is wrong %d expected 3", len(ins.Instructions))
	}
	expected := &Instruction{
		Start: Start{Max: 1, Interval: 200},
	}
	if reflect.DeepEqual(ins.Instructions[0], expected) {
		log.Fatalf("instructions size is wrong %v expected %v", ins.Instructions[0], expected)
	}

	expected = &Instruction{
		Float: Float{Rate: 1.0, Duration: 110},
	}
	if reflect.DeepEqual(ins.Instructions[1], expected) {
		log.Fatalf("instructions size is wrong %v expected %v", ins.Instructions[1], expected)
	}

	expected = &Instruction{
		ExpectRunning: ExpectRunning{Symbol: "<", Amount: 10},
	}
	if reflect.DeepEqual(ins.Instructions[2], expected) {
		log.Fatalf("instructions size is wrong %v expected %v", ins.Instructions[2], expected)
	}

	expected = &Instruction{
		Sleep: 200,
	}
	if reflect.DeepEqual(ins.Instructions[3], expected) {
		log.Fatalf("instructions size is wrong %v expected %v", ins.Instructions[3], expected)
	}

	expected = &Instruction{
		Stop: "stop-all",
	}
	if reflect.DeepEqual(ins.Instructions[4], expected) {
		log.Fatalf("instructions size is wrong %v expected %v", ins.Instructions[4], expected)
	}
}

func TestRawInstructionsDefinition(t *testing.T) {
	rawInstructions := "(sleep 1) (start 2700 100) (sleep 700) (stop-all)"
	def, err := BenchmarkDefByRawInstructions(rawInstructions, 1)
	if err != nil {
		log.Fatalf("unable to parse the raw-instructions test definition: %v", err)
	}

	if len(def.Instructions) != 4 {
		log.Fatalf("instructions size is wrong %d expected 4", len(def.Instructions))
	}

	if def.InstanceGroupSize != 1 {
		log.Fatalf("instance group size is wrong %d expected 1", def.InstanceGroupSize)
	}

	if def.Instructions[3].Stop != "stop-all" {
		log.Fatalf("wrong stop command %s", def.Instructions[3].Stop)
	}

	if def.Instructions[2].Sleep != 700 {
		log.Fatalf("wrong sleep command %d", def.Instructions[2].Sleep)
	}

	if def.Instructions[0].Sleep != 1 {
		log.Fatalf("wrong sleep command %d", def.Instructions[0].Sleep)
	}

	if def.Instructions[1].Start.Max != 2700 && def.Instructions[1].Start.Interval != 100 {
		log.Fatalf("wrong start command %v", def.Instructions[1].Start)
	}
}
