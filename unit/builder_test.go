package unit

import (
	"log"
	"testing"

	"github.com/giantswarm/nomi/definition"
)

func TestBuilder(t *testing.T) {
	// check basic configuration
	app := definition.Application{}
	builder, err := NewBuilder(app, 1, "127.0.0.1:54541", false, false)
	if err != nil {
		log.Fatal(err)
	}
	units1 := builder.MakeUnitChain("1")
	if units1[0].Name != "beaconX-0@1.service" {
		log.Fatalf("wrong unit name expected 'beaconX-0@1.service' got: %s", units1[0].Name)
	}
	options1 := units1[0].Options
	if options1[1].Name != "ExecStart" && options1[1].Value != "/bin/sh -c 'sleep 90000'" {
		log.Fatalf("wrong options name and value expected 'ExecStart' '/bin/sh -c 'sleep 90000'' got: %s %s", options1[1].Name, options1[1].Value)
	}

	// check Docker configuration
	builder, err = NewBuilder(app, 1, "127.0.0.1:54541", true, false)
	units2 := builder.MakeUnitChain("1")
	if units2[0].Name != "beaconX-0@1.service" {
		log.Fatalf("wrong unit name expected 'beaconX-0@1.service' got: %s", units2[0].Name)
	}
	options2 := units2[0].Options
	if options2[1].Name != "ExecStartPre" && options2[1].Value != "-/bin/bash -c '/usr/bin/docker rm -f %p-%i'" {
		log.Fatalf("wrong options name and value expected 'ExecStart' '-/bin/bash -c '/usr/bin/docker rm -f %p-%i'' got: %s %s", options2[1].Name, options2[1].Value)
	}

	// check rkt configuration
	builder, err = NewBuilder(app, 1, "127.0.0.1:54541", false, true)
	units3 := builder.MakeUnitChain("1")
	if units3[0].Name != "beaconX-0@1.service" {
		log.Fatalf("wrong unit name expected 'beaconX-0@1.service' got: %s", units3[0].Name)
	}
	options3 := units3[0].Options
	if options3[3].Name != "KillMode" && options3[3].Value != "Mixed" {
		log.Fatalf("wrong options name and value expected 'KillMode' 'Mixed' got: %s %s", options3[3].Name, options3[3].Value)
	}
}
