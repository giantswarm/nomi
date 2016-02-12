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

func (f *fleetClient) ListUnits() ([]*schema.Unit, error) {
	return f.api.Units()
}

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

func (f *fleetClient) Stop(unitName string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.api.SetUnitTargetState(unitName, "loaded")
}

func (f *fleetClient) Unload(unitName string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.api.SetUnitTargetState(unitName, "inactive")
}

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
