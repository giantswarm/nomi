// This unit package generates all the require unit files
package unit

import (
	"fmt"

	"github.com/coreos/fleet/schema"
)

const beaconUnitPrefix = "beaconX"

// MakeStatsDumper creates fleemer specific units to collect metrics in each host
func MakeStatsDumper(name, cmd, statsEndpoint, listenAddr string) schema.Unit {
	return schema.Unit{
		Name: beaconUnitPrefix + "-stats-dumper-" + name + ".service",
		Options: []*schema.UnitOption{
			{
				Section: "Service",
				Name:    "ExecStart",
				Value:   "/bin/bash -c 'while : ; do " + cmd + "| curl -s -X POST -d @- http://" + listenAddr + "/stats/" + statsEndpoint + " ; done'",
			},
			{
				Section: "X-Fleet",
				Name:    "Global",
				Value:   "true",
			},
		},
	}
}

// MakeUnitChain creates the unit files of the benchmark units.
func MakeUnitChain(id, listenAddr string, chainLength int) []schema.Unit {
	unitsList := []schema.Unit{}

	// 3 dependsOn 2; 2 dependsOn 1; 1 dependsOn 0
	// 0 after 1; 1 after 2 ; 2 after 3

	for i := chainLength - 1; i >= 0; i-- {
		unit := schema.Unit{
			Name: fmt.Sprintf("%s-%d@%s.service", beaconUnitPrefix, i, id),
			Options: []*schema.UnitOption{
				{
					Section: "Service",
					Name:    "ExecStartPre",
					Value:   "/bin/sh -c '/usr/bin/curl -s http://" + listenAddr + "/hello/%i'",
				},
				{
					Section: "Service",
					Name:    "ExecStart",
					Value:   "/bin/sh -c 'sleep 90000'",
					//Value:   "/bin/sh -c 'while [[ -n $(curl -s http://" + listenAddr + "/alive/%i) ]] ; do sleep 1 ; done'",
				},
				{
					Section: "Service",
					Name:    "TimeoutStopSec",
					Value:   "1s",
				},
				{
					Section: "Service",
					Name:    "KillSignal",
					Value:   "SIGKILL",
				},
				{
					Section: "Service",
					Name:    "ExecStopPost",
					Value:   "/usr/bin/curl -s http://" + listenAddr + "/bye/%i",
				},
			},
		}

		if i > 0 {
			depName := fmt.Sprintf("%s-%d@%s.service", beaconUnitPrefix, i-1, id)
			unit.Options = append(unit.Options,
				// &schema.UnitOption{
				// 	Section: "Unit",
				// 	Name:    "Before",
				// 	Value:   depName,
				// },
				&schema.UnitOption{
					Section: "X-Fleet",
					Name:    "MachineOf",
					Value:   depName,
				},
			)
		}
		unitsList = append(unitsList, unit)
	}

	if len(unitsList) > 1 {
		for i := 0; i < (len(unitsList) - 1); i++ {
			depName := unitsList[i+1].Name
			unitsList[i].Options = append(unitsList[i].Options,
				&schema.UnitOption{
					Section: "Unit",
					Name:    "Before",
					Value:   depName,
				},
				&schema.UnitOption{
					Section: "Unit",
					Name:    "BindTo",
					Value:   depName,
				},
			)
		}
	}

	return unitsList
}
