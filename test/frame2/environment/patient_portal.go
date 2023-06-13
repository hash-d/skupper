package environment

import (
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/deploy"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
)

// A Patient Portal deployment on pub1 (frontend), prv1 (DB) and prv2 (payment),
// on the default topology
type PatientPortalDefault struct {
	Name         string
	AutoTearDown bool

	// If true, console will be enabled on prv1
	EnableConsole bool

	// Return

	TopoReturn topology.Basic
	frame2.DefaultRunDealer
}

func (p *PatientPortalDefault) Execute() error {

	name := p.Name
	if name == "" {
		name = "patient-portal"
	}

	baseRunner := base.ClusterTestRunnerBase{}

	var topoSimplest topology.Basic
	topoSimplest = &topologies.Simplest{
		Name:             name,
		TestRunnerBase:   &baseRunner,
		ConsoleOnPrivate: p.EnableConsole,
	}

	p.TopoReturn = topoSimplest

	execute := &frame2.Phase{
		Runner: p.GetRunner(),
		Doc:    "Default Patient Portal deployment",
		MainSteps: []frame2.Step{
			{
				Modify: &PatientPortal{
					Topology:      &topoSimplest,
					AutoTearDown:  p.AutoTearDown,
					SkupperExpose: true,
				},
			},
		},
	}

	return execute.Run()
}

type PatientPortal struct {
	Topology      *topology.Basic
	AutoTearDown  bool
	SkupperExpose bool

	// If true, console will be enabled on prv1
	EnableConsole bool

	frame2.DefaultRunDealer
}

func (p PatientPortal) Execute() error {
	topo := topology.TopologyBuild{
		Topology:     p.Topology,
		AutoTearDown: p.AutoTearDown,
	}

	execute := frame2.Phase{
		Runner: p.GetRunner(),
		Doc:    "Deploy a Patient Portal environment",
		MainSteps: []frame2.Step{
			{
				Modify: &topo,
			}, {
				Modify: &deploy.PatientPortal{
					Topology:      p.Topology,
					SkupperExpose: p.SkupperExpose,
				},
			},
		},
	}
	return execute.Run()
}
