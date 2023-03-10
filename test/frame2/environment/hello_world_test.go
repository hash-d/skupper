package environment

import (
	"fmt"
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
)

func TestHelloWorld(t *testing.T) {

	testRunnerBase := base.ClusterTestRunnerBase{}
	runner := frame2.Run{T: t}

	var topologyN topology.Basic
	topologyN = &topologies.N{
		Name:           "hello-n",
		TestRunnerBase: &testRunnerBase,
	}

	prepareTopology := frame2.Phase{
		Runner: &runner,
		Setup: []frame2.Step{
			{
				Modify: topologyN,
			}, {
				Modify: execute.Print{
					Message: fmt.Sprintf("topologyN: %#v", &topologyN),
				},
			},
		},
	}
	prepareTopology.Run()

	deployApp := frame2.Phase{
		Runner: &runner,
		Setup: []frame2.Step{
			{
				Modify: HelloWorld{
					Runner:       &runner,
					Topology:     &topologyN,
					AutoTearDown: true,
				},
			},
		},
	}
	deployApp.Run()

}
