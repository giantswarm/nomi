// This definition package implements the transformation of a YAML definition
// to a nomi benchmark
package definition

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/giantswarm/nomi/log"
)

type StopCommand string
type ExpectRunningSymbol string

const (
	instructionStart         = "start"
	instructionFloat         = "float"
	instructionExpectRunning = "expect-running"
	instructionSleep         = "sleep"
	instructionStop          = "stop"

	StopAll StopCommand = "stop-all"

	Lower   ExpectRunningSymbol = "<"
	Greater ExpectRunningSymbol = ">"
)

type Start struct {
	Max      int `yaml:"max"`
	Interval int `yaml:"interval"`
}
type Float struct {
	Rate     float64 `yaml:"rate"`
	Duration int     `yaml:"duration"`
}
type ExpectRunning struct {
	Symbol ExpectRunningSymbol `yaml:"symbol"`
	Amount int                 `yaml:"amount"`
}

type Instruction struct {
	Start         Start
	Float         Float
	ExpectRunning ExpectRunning
	Sleep         int         `yaml:"sleep"`
	Stop          StopCommand `yaml:"stop"`
}
type Instructions []Instruction

type Application struct {
	Name    string
	Image   string
	Type    string
	Network string
	Volumes Volumes
	Ports   []int
	Args    []string
	Envs    map[string]string
}

type Volumes []Volume

type Volume struct {
	Source string
	Target string
}

type BenchmarkDef struct {
	Application       Application
	Instructions      Instructions
	InstanceGroupSize int `yaml:"instancegroup-size"`
}

// BenchmarkDefByFile procudes a benchmark definition out of a YAML file
// Return a benchmark definition object and error
func BenchmarkDefByFile(filePath string) (BenchmarkDef, error) {
	def, err := parseBenchmarkDef(filePath)
	if err != nil {
		log.Logger().Fatalf("error when parsing the benchmark definition %v", err)
	}
	return def, err
}

// BenchmarkDefByRawInstructions creates a benchmark definition using raw
// instructions and instance group size
// Return a benchmark definition and error
func BenchmarkDefByRawInstructions(instructions string, igSize int) (BenchmarkDef, error) {
	re := regexp.MustCompile(`\(([^\)]+)\)`)
	parsed := re.FindAllStringSubmatch(instructions, -1)

	def := BenchmarkDef{}
	def.Instructions = make([]Instruction, 0)

	def.InstanceGroupSize = igSize
	for _, cmdString := range parsed {
		if len(cmdString) < 2 {
			continue
		}
		splitted := strings.Fields(cmdString[1])
		if len(splitted) == 0 {
			continue
		}

		cmd := splitted[0]
		args := []string{}
		if len(splitted) > 1 {
			args = splitted[1:]
		}

		switch cmd {
		case instructionStart:
			if len(args) != 2 {
				log.Logger().Fatalf("start requires 2 arguments: max and time between starts. eg: (start 10 100ms)")
			}

			max, err := strconv.Atoi(args[0])
			if err != nil {
				log.Logger().Fatalf("%v", err)
			}
			var interval int
			interval, err = strconv.Atoi(args[1])
			if err != nil {
				log.Logger().Fatalf("%v", err)
			}

			def.Instructions = append(def.Instructions, Instruction{
				Start: Start{
					Max:      max,
					Interval: interval,
				},
			})
			break
		case instructionFloat:
			if len(args) != 2 {
				log.Logger().Fatalf("float requires 2 arguments: rate and duration")
			}
			rate, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				log.Logger().Fatalf("%v", err)
			}
			var duration int
			duration, err = strconv.Atoi(args[1])
			if err != nil {
				log.Logger().Fatalf("%v", err)
			}

			def.Instructions = append(def.Instructions, Instruction{
				Float: Float{
					Rate:     rate,
					Duration: duration,
				},
			})
			break
		case instructionSleep:
			timeout, err := strconv.Atoi(args[0])
			if err != nil {
				log.Logger().Fatalf("%v", err)
			}

			def.Instructions = append(def.Instructions, Instruction{
				Sleep: timeout,
			})
			break
		case instructionExpectRunning:
			if len(args) != 2 {
				log.Logger().Fatal("expect-running requires 2 arguments: [><] int")
			}

			qty, err := strconv.Atoi(args[1])
			if err != nil {
				log.Logger().Fatalf("%v", err)
			}
			symbol := ExpectRunningSymbol(args[0])
			if symbol != Lower && symbol != Greater {
				log.Logger().Fatalf("expect-running comparator has to be > or <")
			}

			def.Instructions = append(def.Instructions, Instruction{
				ExpectRunning: ExpectRunning{
					Symbol: symbol,
					Amount: qty,
				},
			})
			break
		case "stop-all":
			def.Instructions = append(def.Instructions, Instruction{
				Stop: StopCommand(cmd),
			})
			break
		}
	}

	return def, nil
}

func parseBenchmarkDef(filePath string) (BenchmarkDef, error) {
	filename, _ := filepath.Abs(filePath)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Logger().Fatalf("unable to read yaml file %v", err)
	}

	def := BenchmarkDef{}
	if err := yaml.Unmarshal(yamlFile, &def); err != nil {
		log.Logger().Fatalf("unable to parse the yaml test definition: %v", err)
	}

	if !validateDefinition(def) {
		log.Logger().Fatal("benchmark definition contains wrong values")
	}

	return def, nil
}

func validateDefinition(benchmark BenchmarkDef) bool {
	if benchmark.InstanceGroupSize <= 0 {
		log.Logger().Fatal("instance group size has to be greater or equal to 1")
	}
	emptyInstruction := &Instruction{}

	for _, instruction := range benchmark.Instructions {
		if instruction.Start != emptyInstruction.Start && (instruction.Start.Max == 0 || instruction.Start.Interval < 0) {
			log.Logger().Errorf("wrong values for the start operation of instruction %v", instruction)
			return false
		}
		if instruction.Float != emptyInstruction.Float && (instruction.Float.Rate <= 0 || instruction.Float.Duration <= 0) {
			log.Logger().Errorf("wrong values for the start operation of instruction %v", instruction)
			return false
		}
		if instruction.ExpectRunning != emptyInstruction.ExpectRunning &&
			((instruction.ExpectRunning.Symbol != Lower && instruction.ExpectRunning.Symbol != Greater) || instruction.ExpectRunning.Amount < 0) {
			log.Logger().Errorf("wrong values for the expect-running operation of instruction %v", instruction)
			return false
		}
		if instruction.Sleep < 0 {
			log.Logger().Errorf("wrong values for the sleep operation of instruction %v", instruction)
			return false
		}
		if instruction.Stop != "" && instruction.Stop != StopAll {
			log.Logger().Errorf("wrong values for the stop operation of instruction %v", instruction)
			return false
		}
	}

	return true
}
