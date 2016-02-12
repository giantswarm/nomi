package unit

import "time"

type UnitState struct {
	startRequestTime time.Time
	actualStartTime  time.Time
	stopRequestTime  time.Time
	actualStopTime   time.Time
}
