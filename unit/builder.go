// This unit package generates all the require unit files
package unit

import (
	"fmt"
	"strings"

	"github.com/coreos/fleet/schema"

	"github.com/giantswarm/nomi/definition"
)

const (
	beaconUnitPrefix = "beaconX"

	rktTestImage = "docker://giantswarm/alpine-curl"
)

type Builder struct {
	unitPrefix        string
	listenAddr        string
	useDockerService  bool
	useRktService     bool
	app               definition.Application
	instanceGroupSize int
}

func NewBuilder(app definition.Application, instanceGroupSize int, listenAddr string, useDockerService bool, useRktService bool) (*Builder, error) {
	unitPrefix := app.Name
	if unitPrefix == "" {
		unitPrefix = beaconUnitPrefix
	}
	return &Builder{
		unitPrefix:        unitPrefix,
		instanceGroupSize: instanceGroupSize,
		listenAddr:        listenAddr,
		useDockerService:  useDockerService,
		useRktService:     useRktService,
		app:               app,
	}, nil
}

func (b *Builder) GetUnitPrefix() string {
	prefix := beaconUnitPrefix
	if b.app.Name != "" {
		prefix = b.app.Name
	}
	return prefix
}

// MakeStatsDumper creates fleemer specific units to collect metrics in each host
func (b *Builder) MakeStatsDumper(name, cmd, statsEndpoint string) schema.Unit {
	prefix := beaconUnitPrefix
	if b.app.Name != "" {
		prefix = b.app.Name
	}
	return schema.Unit{
		Name: prefix + "-stats-dumper-" + name + ".service",
		Options: []*schema.UnitOption{
			{
				Section: "Service",
				Name:    "ExecStart",
				Value:   "/bin/bash -c 'while : ; do " + cmd + "| curl -s -X POST -d @- http://" + b.listenAddr + "/stats/" + statsEndpoint + " ; done'",
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
func (b *Builder) MakeUnitChain(id string) []schema.Unit {
	unitsList := []schema.Unit{}

	// 3 dependsOn 2; 2 dependsOn 1; 1 dependsOn 0
	// 0 after 1; 1 after 2 ; 2 after 3

	for i := b.instanceGroupSize - 1; i >= 0; i-- {
		name := fmt.Sprintf("%s-%d@%s.service", b.unitPrefix, i, id)
		unit := schema.Unit{
			Name: name,
		}
		if b.useDockerService || b.app.Type == "docker" {
			unit.Options = b.buildDockerService()
		} else if b.useRktService || b.app.Type == "rkt" {
			unit.Options = b.buildRktService()
		} else {
			unit.Options = b.buildDefaultService()
		}

		if i > 0 {
			depName := fmt.Sprintf("%s-%d@%s.service", b.unitPrefix, i-1, id)
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

func (b *Builder) buildDockerService() []*schema.UnitOption {
	dockerExec := "/usr/bin/docker run --net=host --rm --name %p-%i giantswarm/alpine-curl sleep 90000"
	if b.app.Image != "" {
		ports := ""
		envs := ""
		vols := ""
		for _, port := range b.app.Ports {
			ports = ports + fmt.Sprintf(" -p :%d", port)
		}
		for _, vol := range b.app.Volumes {
			vols = vols + fmt.Sprintf(" -v %s:%s", vol, vol)
		}
		for key, value := range b.app.Envs {
			envs = envs + fmt.Sprintf(" -e %s=%s", key, value)
		}
		dockerExec = "/usr/bin/docker run --net=host --rm" + vols + ports + envs + " --name %p-%i " + b.app.Image + " " + strings.Join(b.app.Args[:], " ")
	}

	unit := []*schema.UnitOption{
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
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + b.listenAddr + "/hello/%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStart",
			Value:   dockerExec,
		},
		{
			Section: "Service",
			Name:    "ExecStop",
			Value:   "-/bin/bash -c '/usr/bin/docker kill %p-%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "/usr/bin/curl -s http://" + b.listenAddr + "/bye/%i",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "-/bin/bash -c '/usr/bin/docker rm -f %p-%i'",
		},
	}

	return unit
}

func (b *Builder) buildRktService() []*schema.UnitOption {
	rktExec := "/usr/bin/rkt --uuid-file-save=/run/rkt-uuids/%p-%i --insecure-skip-verify run --net=host " + rktTestImage + " --exec=/bin/sleep -- 90000"
	if b.app.Image != "" {
		ports := ""
		envs := ""
		vols := ""
		args := ""
		for _, port := range b.app.Ports {
			if strings.Contains(b.app.Image, "docker://") {
				ports = ports + fmt.Sprintf(" --port=%d-tcp:%d", port, port)
			} else {
				ports = ports + fmt.Sprintf(" --port=http:%d", port)
			}
		}
		for id, vol := range b.app.Volumes {
			vols = vols + fmt.Sprintf(" --volume=vol%d,kind=host,source=%s --mount=volume=vol%d,target=%s", id, vol, id, vol)
		}
		for key, value := range b.app.Envs {
			envs = envs + fmt.Sprintf(" --set-env=%s=%s", key, value)
		}
		if len(b.app.Args[:]) > 0 {
			args = "--exec=" + strings.Join(b.app.Args[:], " -- ")
		}
		rktExec = "/usr/bin/rkt --uuid-file-save=/run/rkt-uuids/%p-%i --insecure-skip-verify run --net=host" + vols + ports + envs + " " + b.app.Image + " " + args
	}

	unit := []*schema.UnitOption{
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/usr/bin/mkdir -p /run/rkt-uuids",
		},
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + b.listenAddr + "/hello/%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStart",
			Value:   rktExec,
		},
		{
			Section: "Service",
			Name:    "KillMode",
			Value:   "mixed",
		},
		{
			Section: "Service",
			Name:    "ExecStop",
			Value:   "/usr/bin/curl -s http://" + b.listenAddr + "/bye/%i",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "/usr/bin/rkt rm --uuid-file=/run/rkt-uuids/%p-%i",
		},
	}

	return unit
}

func (b *Builder) buildDefaultService() []*schema.UnitOption {
	return []*schema.UnitOption{
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + b.listenAddr + "/hello/%i'",
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
			Value:   "/usr/bin/curl -s http://" + b.listenAddr + "/bye/%i",
		},
	}
}
