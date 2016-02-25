package unit

import (
	"log"
	"testing"
)

func TestBuilder(t *testing.T) {
	// check basic configuration
	units1 := MakeUnitChain("1", "127.0.0.1:54541", 1, false, false)
	if units1[0].Name != "beaconX-0@1.service" {
		log.Fatalf("wrong unit name expected 'beaconX-0@1.service' got: %s", units1[0].Name)
	}
	options1 := units1[0].Options
	if options1[1].Name != "ExecStart" && options1[1].Value != "/bin/sh -c 'sleep 90000'" {
		log.Fatalf("wrong options name and value expected 'ExecStart' '/bin/sh -c 'sleep 90000'' got: %s %s", options1[1].Name, options1[1].Value)
	}

	// check Docker configuration
	units2 := MakeUnitChain("1", "127.0.0.1:54541", 1, true, false)
	if units2[0].Name != "beaconX-0@1.service" {
		log.Fatalf("wrong unit name expected 'beaconX-0@1.service' got: %s", units2[0].Name)
	}
	options2 := units2[0].Options
	if options2[1].Name != "ExecStartPre" && options2[1].Value != "-/bin/bash -c '/usr/bin/docker rm -f %p-%i'" {
		log.Fatalf("wrong options name and value expected 'ExecStart' '-/bin/bash -c '/usr/bin/docker rm -f %p-%i'' got: %s %s", options2[1].Name, options2[1].Value)
	}

	// check rkt configuration
	units3 := MakeUnitChain("1", "127.0.0.1:54541", 1, false, true)
	if units3[0].Name != "beaconX-0@1.service" {
		log.Fatalf("wrong unit name expected 'beaconX-0@1.service' got: %s", units3[0].Name)
	}
	options3 := units3[0].Options
	if options3[3].Name != "KillMode" && options3[3].Value != "Mixed" {
		log.Fatalf("wrong options name and value expected 'KillMode' 'Mixed' got: %s %s", options3[3].Name, options3[3].Value)
	}
}
