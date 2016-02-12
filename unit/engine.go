package unit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/giantswarm/fleemmer/definition"
)

type stats []statsLine

type event struct {
	Cmd       string
	Args      []string
	StartTime float64
	EndTime   float64
}

type UnitEngine struct {
	benchmark definition.BenchmarkDef

	SpawnFunc func(string) error
	StopFunc  func(string) error

	startingUnits map[string]UnitState
	runningUnits  map[string]UnitState
	stoppingUnits map[string]UnitState
	stoppedUnits  map[string]UnitState

	eventLog []event

	startedStats stats
	stoppedStats stats

	machineStats map[string][]processStatsLine

	startTime time.Time

	mu *sync.Mutex
}

type processStatsLine struct {
	Process   string
	TimeStamp float64
	CPUUsage  float64
	RSS       int
}

type statsLine struct {
	ID             string
	StartTime      float64
	CompletionTime float64
	Delay          float64
	StartingCount  int
	RunningCount   int
	StoppingCount  int
	StoppedCount   int
}

func (l statsLine) toSSV() string {
	return fmt.Sprintf("%s %f %f %d %d %d %d",
		l.ID,
		l.StartTime,
		l.Delay,
		l.StartingCount,
		l.RunningCount,
		l.StoppingCount,
		l.StoppedCount)
}

func NewEngine(def definition.BenchmarkDef) (*UnitEngine, error) {
	return &UnitEngine{
		mu:            new(sync.Mutex),
		benchmark:     def,
		startingUnits: map[string]UnitState{},
		runningUnits:  map[string]UnitState{},
		stoppingUnits: map[string]UnitState{},
		stoppedUnits:  map[string]UnitState{},
		startedStats:  stats{},
		stoppedStats:  stats{},
		eventLog:      []event{},
		machineStats:  map[string][]processStatsLine{},
	}, nil
}

func (e *UnitEngine) InstanceGroupSize() int {
	return e.benchmark.InstanceGroupSize
}

func (e *UnitEngine) Run() {
	defer e.stopAll()
	var (
		emptyStart         definition.Start
		emptyFloat         definition.Float
		emptyExpectRunning definition.ExpectRunning
	)
	e.startTime = time.Now()

	for _, instruction := range e.benchmark.Instructions {
		if instruction.Start != emptyStart {
			go func(obj definition.Start) {
				startTime := time.Now()
				e.start(obj)
				e.logCommand("start", []string{fmt.Sprintf("%d", instruction.Start.Max), fmt.Sprintf("%d", instruction.Start.Interval)}, startTime, time.Now())
			}(instruction.Start)
		}
		if instruction.Float != emptyFloat {
			startTime := time.Now()
			e.float(instruction.Float)
			e.logCommand("float", []string{fmt.Sprintf("%d", instruction.Float.Rate), fmt.Sprintf("%v", instruction.Float.Duration)}, startTime, time.Now())
		}
		if instruction.Sleep != 0 {
			startTime := time.Now()
			time.Sleep(time.Duration(instruction.Sleep) * time.Second)
			e.logCommand("sleep", []string{fmt.Sprintf("%d", instruction.Sleep)}, startTime, time.Now())
		}
		if instruction.ExpectRunning != emptyExpectRunning {
			startTime := time.Now()
			e.expectRunning(instruction.ExpectRunning)
			e.logCommand("expect-running", []string{fmt.Sprintf("%s", instruction.ExpectRunning.Symbol), fmt.Sprintf("%d", instruction.ExpectRunning.Amount)}, startTime, time.Now())
		}
		if instruction.Stop != "" {
			startTime := time.Now()
			e.stopAll()
			e.logCommand("stop-all", []string{fmt.Sprintf("%s", instruction.Stop)}, startTime, time.Now())
		}
	}
}

func (e *UnitEngine) MarkUnitRunning(id string) time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	state, exists := e.startingUnits[id]
	if !exists {
		glog.Errorf("unit %s cannot be found in the starting pool\n", id)
		return time.Duration(0)
	}
	delete(e.startingUnits, id)
	state.actualStartTime = time.Now()
	e.runningUnits[id] = state
	e.startedStats = append(e.startedStats, e.genStatsLine(id, state.actualStartTime.Sub(state.startRequestTime)))
	return state.actualStartTime.Sub(state.startRequestTime)
}

func (e *UnitEngine) MarkUnitStopped(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	state, exists := e.stoppingUnits[id]
	if !exists {
		glog.Errorf("unit %s cannot be found in the stopping pool\n", id)
		return
	}
	delete(e.stoppingUnits, id)
	state.actualStopTime = time.Now()
	e.stoppedUnits[id] = state
	e.stoppedStats = append(e.stoppedStats, e.genStatsLine(id, state.actualStopTime.Sub(state.stopRequestTime)))
}

type Stats struct {
	Start        stats
	Stop         stats
	Script       string
	EventLog     []event
	MachineStats map[string][]processStatsLine
}

func (e *UnitEngine) Stats() Stats {
	return Stats{
		Start:        e.startedStats,
		Stop:         e.stoppedStats,
		EventLog:     e.eventLog,
		MachineStats: e.machineStats,
	}
}

func (e *UnitEngine) DumpProcessStats(statsid, hostname string, cpuusage float64, rss int) {
	statsLine := processStatsLine{
		Process:   statsid,
		CPUUsage:  cpuusage,
		RSS:       rss,
		TimeStamp: time.Now().Sub(e.startTime).Seconds(),
	}

	if _, exists := e.machineStats[hostname]; exists {
		e.machineStats[hostname] = append(e.machineStats[hostname], statsLine)
	} else {
		e.machineStats[hostname] = []processStatsLine{statsLine}
	}
}

func (e *UnitEngine) logCommand(cmd string, args []string, start time.Time, end time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.eventLog = append(e.eventLog,
		event{
			Cmd:       cmd,
			Args:      args,
			StartTime: start.Sub(e.startTime).Seconds(),
			EndTime:   end.Sub(e.startTime).Seconds(),
		})
}

func (e *UnitEngine) expectRunning(obj definition.ExpectRunning) {
	for {
		running := len(e.runningUnits)
		if obj.Symbol == ">" && running > obj.Amount {
			return
		}
		if obj.Symbol == "<" && running < obj.Amount {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (e *UnitEngine) float(obj definition.Float) {
	glog.V(2).Infoln("float instruction not implemented yet...")
	//duration := time.Duration(obj.Duration) * time.Millisecond
	//glog.V(2).Infof("floating for %s\n", duration)
	//time.Sleep(duration)
}

func genRandomID() string {
	c := 5
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (e *UnitEngine) start(obj definition.Start) {
	spawnUnit := func() {
		newID := genRandomID()
		e.startingUnits[newID] = UnitState{startRequestTime: time.Now()}
		e.SpawnFunc(newID)
	}

	wg := new(sync.WaitGroup)
	for spawned := 0; spawned < obj.Max; spawned++ {
		wg.Add(1)
		go func() {
			spawnUnit()
			wg.Done()
		}()
		time.Sleep(time.Duration(obj.Interval) * time.Millisecond)
	}
	wg.Wait()
}

func (e *UnitEngine) stopUnit(id string, state UnitState) {
	e.mu.Lock()
	newState := state
	newState.stopRequestTime = time.Now()
	glog.V(3).Infoln("marking beacon as to be deleted: " + id)
	e.stoppingUnits[id] = newState
	e.mu.Unlock()
	err := e.StopFunc(id)
	if err != nil {
		glog.Warningln(err)
	}
}

func (e *UnitEngine) stopAll() {
	wg := new(sync.WaitGroup)
	for id, state := range e.startingUnits {
		wg.Add(1)
		go func(id string, state UnitState) {
			e.stopUnit(id, state)
			wg.Done()
		}(id, state)
		e.mu.Lock()
		delete(e.startingUnits, id)
		e.mu.Unlock()

	}

	for id, state := range e.runningUnits {
		wg.Add(1)
		go func(id string, state UnitState) {
			e.stopUnit(id, state)
			wg.Done()
		}(id, state)
		e.mu.Lock()
		delete(e.runningUnits, id)
		e.mu.Unlock()
	}
	wg.Wait()
}

func (e *UnitEngine) genStatsLine(id string, delay time.Duration) statsLine {
	startTime := time.Now().Add(-delay)

	return statsLine{
		ID:             id,
		StartTime:      startTime.Sub(e.startTime).Seconds(),
		CompletionTime: startTime.Add(delay).Sub(e.startTime).Seconds(),
		Delay:          delay.Seconds(),
		StartingCount:  len(e.startingUnits),
		RunningCount:   len(e.runningUnits),
		StoppingCount:  len(e.stoppingUnits),
		StoppedCount:   len(e.stoppedUnits),
	}
}
