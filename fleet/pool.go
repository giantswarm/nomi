package fleet

import "github.com/coreos/fleet/schema"

type fleetPool struct {
	fleets   []fleetAPI
	nextConn int
}

// NewFleetPool creats a pool of connections to a remote fleet API
func NewFleetPool(size int) fleetAPI {
	fleets := make([]fleetAPI, size)

	for i := 0; i < size; i++ {
		fleets[i] = newFleet()
	}

	return &fleetPool{fleets, 0}
}

// StartUnitGroup starts an unit group
func (p *fleetPool) StartUnitGroup(units []schema.Unit) error {
	return p.getFleetClient().StartUnitGroup(units)
}

// StartUnit starts an unit by passing
func (p *fleetPool) StartUnit(unit schema.Unit) error {
	return p.getFleetClient().StartUnit(unit)
}

// CleanupPrefix cleans all units with a specific prefix
func (p *fleetPool) CleanupPrefix(prefix string) error {
	return p.getFleetClient().CleanupPrefix(prefix)
}

// Stop stops an unit by passing its unit name
func (p *fleetPool) Stop(unitName string) error {
	return p.getFleetClient().Stop(unitName)
}

// Destroy destroys an unit by passing its unit name
func (p *fleetPool) Destroy(unitName string) error {
	return p.getFleetClient().Destroy(unitName)
}

// Unload unloads an unit by passing its unit name
func (p *fleetPool) Unload(unitName string) error {
	return p.getFleetClient().Unload(unitName)
}

// ListUnits returns the list of units in the fleet cluster
func (p *fleetPool) ListUnits() ([]*schema.Unit, error) {
	return p.getFleetClient().ListUnits()
}

func (p *fleetPool) getFleetClient() fleetAPI {
	f := p.fleets[p.nextConn]
	p.nextConn = (p.nextConn + 1) % len(p.fleets)
	return f
}
