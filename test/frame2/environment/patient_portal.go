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
	Runner       *frame2.Run
	AutoTearDown bool

	// Return

	TopoReturn topology.Basic
}

func (p *PatientPortalDefault) Execute() error {

	name := p.Name
	if name == "" {
		name = "patient-portal"
	}

	baseRunner := base.ClusterTestRunnerBase{}

	var topoSimplest topology.Basic
	topoSimplest = &topologies.Simplest{
		Name:           name,
		TestRunnerBase: &baseRunner,
	}

	p.TopoReturn = topoSimplest

	execute := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: PatientPortal{
					Runner:        p.Runner,
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
	Runner        *frame2.Run // Required for autoTeardown and step logging
	Topology      *topology.Basic
	AutoTearDown  bool
	SkupperExpose bool
}

func (p PatientPortal) Execute() error {
	topo := topology.TopologyBuild{
		Runner:       p.Runner,
		Topology:     p.Topology,
		AutoTearDown: p.AutoTearDown,
	}

	execute := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &topo,
			}, {
				Modify: deploy.PatientPortal{
					Runner:        p.Runner,
					Topology:      p.Topology,
					SkupperExpose: p.SkupperExpose,
				},
			},
		},
	}
	return execute.Run()
}
