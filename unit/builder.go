// This unit package generates all the require unit files
package unit

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/coreos/fleet/schema"
	"github.com/coreos/fleet/unit"

	"github.com/giantswarm/nomi/definition"
	"github.com/giantswarm/nomi/log"
)

const (
	nomiUnitPrefix = "nomi"

	rktTestImage = "docker://giantswarm/alpine-curl"
)

type Builder struct {
	unitPrefix        string
	listenAddr        string
	app               definition.Application
	instanceGroupSize int
	unitFile          *unit.UnitFile
}

func NewBuilder(app definition.Application, instanceGroupSize int, listenAddr string) (*Builder, error) {
	unitPrefix := app.Name
	if unitPrefix == "" {
		unitPrefix = nomiUnitPrefix
	}
	return &Builder{
		unitPrefix:        unitPrefix,
		instanceGroupSize: instanceGroupSize,
		listenAddr:        listenAddr,
		app:               app,
	}, nil
}

func (b *Builder) GetUnitPrefix() string {
	prefix := nomiUnitPrefix
	if b.app.Name != "" {
		prefix = b.app.Name
	}
	return prefix
}

// MakeStatsDumper creates nomi specific units to collect metrics in each host
func (b *Builder) MakeStatsDumper(name, cmd, statsEndpoint string) schema.Unit {
	prefix := nomiUnitPrefix
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
		if b.app.Type == "unitfiles" && b.unitFile != nil {
			unit.Options = b.buildCustomService()
		} else if b.app.Type == "docker" {
			unit.Options = b.buildDockerService()
		} else if b.app.Type == "rkt" {
			unit.Options = b.buildRktService()
		} else {
			unit.Options = b.buildShellService()
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

func (b *Builder) UseCustomUnitFileService(filePath string) error {
	filename, _ := filepath.Abs(filePath)
	unitFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Logger().Errorf("unable to read yaml file %v", err)
		return err
	}
	contents := string(unitFile)
	if strings.Contains(contents, "Global=true") {
		log.Logger().Warningf("Global fleet scheduling option 'Global=true' can cause undesired results")
	}

	b.unitFile, err = unit.NewUnitFile(contents)
	if err != nil {
		log.Logger().Errorf("error creating Unit from %q: %v", contents, err)
		return err
	}
	return nil
}

func (b *Builder) generateDockerRunCmd() string {
	ports := ""
	envs := ""
	vols := ""
	net := ""
	for _, port := range b.app.Ports {
		ports = ports + fmt.Sprintf(" -p :%d", port)
	}
	for _, vol := range b.app.Volumes {
		vols = vols + fmt.Sprintf(" -v %s:%s", vol.Source, vol.Target)
	}
	for key, value := range b.app.Envs {
		envs = envs + fmt.Sprintf(" -e %s=%s", key, value)
	}
	if b.app.Network != "" {
		net = " --net=" + b.app.Network
	}
	return "/usr/bin/docker run --rm" + net + vols + ports + envs + " --name %p-%i " + b.app.Image + " " + strings.Join(b.app.Args[:], " ")
}

func (b *Builder) buildDockerService() []*schema.UnitOption {
	dockerExec := "/usr/bin/docker run --net=host --rm --name %p-%i giantswarm/alpine-curl sleep 90000"
	if b.app.Image != "" {
		dockerExec = b.generateDockerRunCmd()
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

func (b *Builder) generateRktRunCmd() string {
	ports := ""
	envs := ""
	vols := ""
	args := ""
	net := ""
	for _, port := range b.app.Ports {
		if strings.Contains(b.app.Image, "docker://") {
			ports = ports + fmt.Sprintf(" --port=%d-tcp:%d", port, port)
		} else {
			ports = ports + fmt.Sprintf(" --port=http:%d", port)
		}
	}
	// TODO: This notation uses newer versions of rkt 1.0.0, older don't support '--mount'
	for id, vol := range b.app.Volumes {
		vols = vols + fmt.Sprintf(" --volume=vol%d,kind=host,source=%s --mount=volume=vol%d,target=%s", id, vol.Source, id, vol.Target)
	}
	for key, value := range b.app.Envs {
		envs = envs + fmt.Sprintf(" --set-env=%s=%s", key, value)
	}
	if len(b.app.Args[:]) > 0 {
		args = "--exec=" + strings.Join(b.app.Args[:], " -- ")
	}
	if b.app.Network != "" {
		net = " --net=" + b.app.Network
	}
	return "/usr/bin/rkt --uuid-file-save=/run/rkt-uuids/%p-%i --insecure-skip-verify run " + net + vols + ports + envs + " " + b.app.Image + " " + args
}

func (b *Builder) buildRktService() []*schema.UnitOption {
	rktExec := "/usr/bin/rkt --uuid-file-save=/run/rkt-uuids/%p-%i --insecure-skip-verify run --net=host " + rktTestImage + " --exec=/bin/sleep -- 90000"
	if b.app.Image != "" {
		rktExec = b.generateRktRunCmd()
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

func (b *Builder) buildShellService() []*schema.UnitOption {
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

func (b *Builder) buildCustomService() []*schema.UnitOption {
	unitOptions := schema.MapUnitFileToSchemaUnitOptions(b.unitFile)

	nomiNotifiers := []*schema.UnitOption{
		{
			Section: "Service",
			Name:    "ExecStartPre",
			Value:   "/bin/sh -c '/usr/bin/curl -s http://" + b.listenAddr + "/hello/%i'",
		},
		{
			Section: "Service",
			Name:    "ExecStopPost",
			Value:   "/usr/bin/curl -s http://" + b.listenAddr + "/bye/%i",
		},
	}

	unitOptions = append(unitOptions, nomiNotifiers...)

	return unitOptions
}
