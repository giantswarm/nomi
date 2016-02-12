package fleet

import "github.com/coreos/fleet/schema"

type fleetPool struct {
	fleets   []fleetAPI
	nextConn int
}

func NewFleetPool(size int) fleetAPI {
	fleets := make([]fleetAPI, size)

	for i := 0; i < size; i++ {
		fleets[i] = newFleet()
	}

	return &fleetPool{fleets, 0}
}

func (p *fleetPool) StartUnitGroup(units []schema.Unit) error {
	return p.getFleetClient().StartUnitGroup(units)
}

func (p *fleetPool) StartUnit(unit schema.Unit) error {
	return p.getFleetClient().StartUnit(unit)
}

func (p *fleetPool) CleanupPrefix(prefix string) error {
	return p.getFleetClient().CleanupPrefix(prefix)
}

func (p *fleetPool) Stop(unitName string) error {
	return p.getFleetClient().Stop(unitName)
}

func (p *fleetPool) Destroy(unitName string) error {
	return p.getFleetClient().Destroy(unitName)
}

func (p *fleetPool) Unload(unitName string) error {
	return p.getFleetClient().Unload(unitName)
}

func (p *fleetPool) ListUnits() ([]*schema.Unit, error) {
	return p.getFleetClient().ListUnits()
}

func (p *fleetPool) getFleetClient() fleetAPI {
	f := p.fleets[p.nextConn]
	p.nextConn = (p.nextConn + 1) % len(p.fleets)
	return f
}
