package unit

import "time"

// UnitState collects the timestamp of the operations
type UnitState struct {
	startRequestTime time.Time
	actualStartTime  time.Time
	stopRequestTime  time.Time
	actualStopTime   time.Time
}
