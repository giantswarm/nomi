// This fleet package implements all the operation to communicate with a fleet
// cluster
package fleet

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/coreos/fleet/client"
	"github.com/coreos/fleet/schema"
)

// fleetAPI represents all the operations we want to perform to a fleet API
type fleetAPI interface {
	StartUnit(unit schema.Unit) error
	StartUnitGroup(unit []schema.Unit) error
	ListUnits() ([]*schema.Unit, error)
	CleanupPrefix(prefix string) error
	Stop(unitName string) error
	Destroy(unitName string) error
	Unload(unitName string) error
}

type fleetClient struct {
	api client.API
	m   *sync.Mutex
}

// ListUnits returns the list of units in the fleet cluster
func (f *fleetClient) ListUnits() ([]*schema.Unit, error) {
	return f.api.Units()
}

// StartUnitGroup starts an instance group by passing as input argument the instance group
func (f *fleetClient) StartUnitGroup(units []schema.Unit) error {
	f.m.Lock()
	defer f.m.Unlock()

	for _, unit := range units {
		err := f.api.CreateUnit(&unit)
		if err != nil {
			return err
		}
		err = f.api.SetUnitTargetState(unit.Name, "loaded")
		if err != nil {
			return err
		}
	}

	for _, unit := range units {
		err := f.api.SetUnitTargetState(unit.Name, "launched")
		if err != nil {
			return err
		}
	}
	return nil
}

// StartUnit starts a specific unit in the cluster
func (f *fleetClient) StartUnit(unit schema.Unit) error {
	f.m.Lock()
	defer f.m.Unlock()

	err := f.api.CreateUnit(&unit)
	if err != nil {
		return err
	}

	err = f.api.SetUnitTargetState(unit.Name, "loaded")
	if err != nil {
		return err
	}

	err = f.api.SetUnitTargetState(unit.Name, "launched")
	if err != nil {
		return err
	}
	return nil
}

// CleanupPrefix destroys all units with a specific prefix
func (f *fleetClient) CleanupPrefix(prefix string) error {
	f.m.Lock()
	defer f.m.Unlock()
	units, err := f.api.Units()
	if err != nil {
		return err
	}

	for _, unit := range units {
		if strings.HasPrefix(unit.Name, prefix) {
			f.api.DestroyUnit(unit.Name)
		}
	}

	return nil
}

// Stop stops an unit by passing its unit name
func (f *fleetClient) Stop(unitName string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.api.SetUnitTargetState(unitName, "loaded")
}

// Unload unloads an unit by passing its unit name
func (f *fleetClient) Unload(unitName string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.api.SetUnitTargetState(unitName, "inactive")
}

// Destroy destroys an unit by passing its unit name
func (f *fleetClient) Destroy(unitName string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.api.DestroyUnit(unitName)
}

func newFleet() *fleetClient {

	cl := &http.Client{
		Transport: &http.Transport{
			Dial: func(string, string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/fleet.sock")
			},
		},
	}

	url, err := url.Parse("http://localhost/")
	if err != nil {
		panic(err)
	}
	api, err := client.NewHTTPClient(cl, *url)
	if err != nil {
		panic(err)
	}

	return &fleetClient{api, new(sync.Mutex)}
}
