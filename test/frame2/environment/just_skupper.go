package environment

import (
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
)

// A Skupper deployment on pub1 (frontend) and prv1 (backend),
// on the default topology
type JustSkupperDefault struct {
	Name         string
	AutoTearDown bool
	Console      bool

	// Return
	Topo topology.Basic

	frame2.DefaultRunDealer
}

func (j *JustSkupperDefault) Execute() error {

	name := j.Name
	if name == "" {
		name = "just-skupper"
	}

	baseRunner := base.ClusterTestRunnerBase{}

	j.Topo = &topologies.Simplest{
		Name:             name,
		TestRunnerBase:   &baseRunner,
		ConsoleOnPublic:  j.Console,
		ConsoleOnPrivate: j.Console,
	}

	execute := frame2.Phase{
		Runner: j.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &JustSkupper{
					Topology:     &j.Topo,
					AutoTearDown: j.AutoTearDown,
				},
			},
		},
	}

	return execute.Run()
}

// A Skupper deployment on pub1 (frontend) and prv1 (backend),
// on an N topology.
//
// Useful for the simplest multiple link testing.
//
// See topology.N for details on the topology.
type JustSkupperN struct {
}

// As the name says, it's just skupper, connected according to the provided
// topology.  For simpler alternatives, see:
//
//   - environment.JustSkupperSimple
//   - environment.JustSkupperN
//   - ...
//   - environment.JustSkupperPlatform is special. It will use
//     whatever topology the current test is asking for, if
//     possible
type JustSkupper struct {
	Topology      *topology.Basic
	AutoTearDown  bool
	SkupperExpose bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (j JustSkupper) Execute() error {
	topo := topology.TopologyBuild{
		Topology:     j.Topology,
		AutoTearDown: j.AutoTearDown,
	}

	execute := frame2.Phase{
		Runner: j.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &topo,
			},
		},
	}
	return execute.Run()
}
