package topology

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

func TestTopologyMap(t *testing.T) {
	runner := base.ClusterTestRunnerBase{}

	pub1 := &TopologyItem{
		Type: Public,
	}

	prv1 := &TopologyItem{
		Type: Private,
		Connections: []*TopologyItem{
			pub1,
		},
	}

	topoMap := []*TopologyItem{
		pub1,
		prv1,
	}

	tests := frame2.TestRun{
		Name: "TestTopology",
		Setup: []frame2.Step{
			{
				Modify: &Topology{
					TopologyMap: TopologyMap{
						Name:           "topo",
						TestRunnerBase: runner,
						Map:            topoMap,
					},
					AutoTearDown: true,
				},
			},
		},
		MainSteps: []frame2.Step{
			{
				Doc: "Show it to me",
				Modify: execute.Print{
					Data: []interface{}{topoMap},
				},
			},
		},
	}

	tests.Run(t)
}
