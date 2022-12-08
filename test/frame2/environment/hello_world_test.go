package environment

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

func TestHelloWorld(t *testing.T) {

	testRunnerBase := base.ClusterTestRunnerBase{}

	topologyN := topology.N{
		Name:           "hello-n",
		TestRunnerBase: &testRunnerBase,
	}

	for i, topo := range []frame2.Executor{&topologyN} {
		err := topo.Execute()
		if err != nil {
			t.Fatalf("Failed to get topology %d(%v): %v", i, topo, err)
		}
	}

	tests := frame2.TestRun{
		Name: "TestHelloWorld",
		Setup: []frame2.Step{
			{
				Modify: HelloWorld{
					TopologyMap: *topologyN.Return,
				},
			},
		},
	}

	tests.Run(t)

}
