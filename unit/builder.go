// This unit package generates all the require unit files
package unit

import (
	"fmt"

	"github.com/coreos/fleet/schema"
)

const (
	beaconUnitPrefix = "beaconX"

	rktTestImageUrl = "https://s3-eu-west-1.amazonaws.com/gs-rkt-test/giantswarm-alpine-curl-latest.aci"
)

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
func MakeUnitChain(id, listenAddr string, chainLength int, useDockerService bool, useRktService bool) []schema.Unit {
	unitsList := []schema.Unit{}

	// 3 dependsOn 2; 2 dependsOn 1; 1 dependsOn 0
	// 0 after 1; 1 after 2 ; 2 after 3

	for i := chainLength - 1; i >= 0; i-- {
		name := fmt.Sprintf("%s-%d@%s.service", beaconUnitPrefix, i, id)
		var unit schema.Unit
		if useDockerService {
			unit = schema.Unit{
				Name:    name,
				Options: buildDockerService(listenAddr),
			}
		} else if useRktService {
			unit = schema.Unit{
				Name:    name,
				Options: buildRktService(listenAddr),
			}
		} else {
			unit = schema.Unit{
				Name:    name,
				Options: buildDefaultService(listenAddr),
			}
		}

		if i > 0 {
			depName := fmt.Sprintf("%s-%d@%s.service", beaconUnitPrefix, i-1, id)
			unit.Options = append(unit.Options,
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

func buildDockerService(listenAddr string) []*schema.UnitOption {
	return []*schema.UnitOption{
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "-/bin/bash -c '/usr/bin/docker kill %p-%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "-/bin/bash -c '/usr/bin/docker rm -f %p-%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + listenAddr + "/hello/%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStart",
			Value:   "/usr/bin/docker run --net=host --rm --name %p-%i giantswarm/alpine-curl sleep 90000",
		},
		{
			Section: "Service",
			Name:    "ExecStop",
			Value:   "-/bin/bash -c '/usr/bin/docker kill %p-%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "/usr/bin/curl -s http://" + listenAddr + "/bye/%i",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "-/bin/bash -c '/usr/bin/docker rm -f %p-%i'",
		},
	}
}

func buildRktService(listenAddr string) []*schema.UnitOption {
	return []*schema.UnitOption{
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/usr/bin/mkdir -p /run/rkt-uuids",
		},
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + listenAddr + "/hello/%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStart",
			Value:   "/usr/bin/rkt --uuid-file-save=/run/rkt-uuids/%p-%i --insecure-skip-verify run --net=host " + rktTestImageUrl + " --exec=/bin/sleep -- 90000",
		},
		{
			Section: "Service",
			Name:    "KillMode",
			Value:   "mixed",
		},
		{
			Section: "Service",
			Name:    "ExecStop",
			Value:   "/usr/bin/curl -s http://" + listenAddr + "/bye/%i",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "/usr/bin/rkt rm --uuid-file=/run/rkt-uuids/%p-%i",
		},
	}
}

func buildDefaultService(listenAddr string) []*schema.UnitOption {
	return []*schema.UnitOption{
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + listenAddr + "/hello/%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStart",
			Value:   "/bin/sh -c 'sleep 90000'",
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
	}
}
