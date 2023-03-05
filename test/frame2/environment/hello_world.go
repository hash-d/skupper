package environment

import (
	"log"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/deploy"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
)

// A Hello World deployment on pub1 (frontend) and prv1 (backend),
// on the default topology
type HelloWorldDefault struct {
	Name   string
	Runner *frame2.Run
}

func (hwd HelloWorldDefault) Execute() error {

	name := hwd.Name
	if name == "" {
		name = "hello-world"
	}

	baseRunner := base.ClusterTestRunnerBase{}

	var topoMap topology.Basic
	topoMap = &topologies.Simplest{
		Name:           name,
		TestRunnerBase: &baseRunner,
	}

	execute := frame2.Phase{
		Runner: hwd.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: HelloWorld{
					Runner:   hwd.Runner,
					Topology: &topoMap,
				},
			},
		},
	}

	return execute.Run()
}

// A Hello World deployment on pub1 (frontend) and prv1 (backend),
// on an N topology.
//
// Useful for the simplest multiple link testing.
//
// See topology.N for details on the topology.
type HelloWorldN struct {
}

// A Hello World deployment, with configurations.  For simpler
// alternatives, see:
//
//   - environment.HelloWorldSimple
//   - environment.HelloWorldN
//   - ...
//   - environment.HelloWorldPlatform is special. It will use
//     whatever topology the current test is asking for, if
//     possible
//
// To use the auto tearDown, make sure to populate the Runner
type HelloWorld struct {
	Runner   *frame2.Run // Required for autoTeardown and step logging
	Topology *topology.Basic
}

func (hw HelloWorld) Execute() error {
	topo := topology.TopologyBuild{
		Runner:       hw.Runner,
		Topology:     hw.Topology,
		AutoTearDown: true,
	}

	execute := frame2.Phase{
		Runner: hw.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: execute.Function{
					Fn: func() error {
						tm, err := (*hw.Topology).GetTopologyMap()
						log.Printf("topo: %+v\nTopology: %+v (%+v)", topo, tm, err)
						return nil
					},
				},
			}, {
				Modify: &topo,
			}, {
				Modify: execute.Function{
					Fn: func() error {
						tm, err := (*hw.Topology).GetTopologyMap()
						log.Printf("topo: %+v\nTopology: %+v (%+v)", topo, tm, err)
						return nil
					},
				},
			}, {
				Modify: deploy.HelloWorld{
					Topology: hw.Topology,
				},
			}, {
				Modify: execute.Function{
					Fn: func() error {
						tm, err := (*hw.Topology).GetTopologyMap()
						log.Printf("topo: %+v\nTopology: %+v (%+v)", topo, tm, err)
						return nil
					},
				},
			},
		},
	}
	return execute.Run()
}
