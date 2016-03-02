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

var dataAppTest = `
application:
 name: helloworld
 image: alpine-curl
 type: docker
 volumes: ["/usr/lib/vol1", "/usr/lib/vol2"]
 ports: [9090, 8080]
 args: ["sleep", "9000"]
 envs:
  x1: value1
  x2: value2
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

func TestApplicationYAMLDefinition(t *testing.T) {
	ins := BenchmarkDef{}

	if err := yaml.Unmarshal([]byte(dataAppTest), &ins); err != nil {
		log.Fatalf("unable to parse the yaml test definition: %v", err)
	}

	if ins.InstanceGroupSize != 1 {
		log.Fatalf("wrong instance group size is wrong %d expected 1", ins.InstanceGroupSize)
	}
	if len(ins.Instructions) != 5 {
		log.Fatalf("instructions size is wrong %d expected 3", len(ins.Instructions))
	}

	expected := &Application{
		Name:    "helloword",
		Image:   "alpine-curl",
		Type:    "docker",
		Args:    []string{"sleep", "9000"},
		Volumes: []string{"/usr/lib/vol1", "/usr/lib/vol2"},
		Ports:   []int{9090, 8080},
		Envs:    make(map[string]string),
	}
	expected.Envs["x1"] = "value1"
	expected.Envs["x2"] = "value2"

	if reflect.DeepEqual(ins.Application.Name, expected.Name) {
		log.Fatalf("application name is wrong %v expected %v", ins.Application.Name, expected.Name)
	}

	if ins.Application.Image != expected.Image {
		log.Fatalf("application image is wrong %v expected %v", ins.Application.Image, expected.Image)
	}

	if ins.Application.Type != expected.Type {
		log.Fatalf("application type is wrong %v expected %v", ins.Application.Type, expected.Type)
	}

	if len(ins.Application.Ports) != len(expected.Ports) && ins.Application.Ports[0] != expected.Ports[0] {
		log.Fatalf("application ports is wrong %v expected %v", ins.Application.Ports, expected.Ports)
	}

	if len(ins.Application.Args) != len(expected.Args) && ins.Application.Args[0] != expected.Args[0] {
		log.Fatalf("application args is wrong %v expected %v", ins.Application.Args, expected.Args)
	}

	if len(ins.Application.Volumes) != len(expected.Volumes) && ins.Application.Volumes[0] != expected.Volumes[0] {
		log.Fatalf("application volumes is wrong %v expected %v", ins.Application.Volumes, expected.Volumes)
	}

	if len(ins.Application.Envs) != len(expected.Envs) && ins.Application.Envs["x1"] != expected.Envs["x1"] {
		log.Fatalf("application environment variable is wrong %v expected %v", ins.Application.Envs["x1"], expected.Envs["x1"])
	}
}
