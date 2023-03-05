// Install Hello World in a 1-1 topology; front-end on pub,
// backend on prv.  Add a new skupper node on a third
// namespace and move part of hello world there.  Once
// good, remove the same from the original namespace (app
// and Skupper).  Validate all good, and move back.
//
// repeat it a few times (or 90% of the alloted test time)
//
// Options:
//
// - remove service first
// - remove link first
// - skupper delete, direct
//
// By default, use a different one each time, but allow
// for selecting a single one
package pingpong

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
)

func TestPingPong(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	var runner = &base.ClusterTestRunnerBase{}

	var topology topology.Basic
	topology = &topologies.V{
		Name:           "pingpong",
		TestRunnerBase: runner,
		EmptyRight:     true,
	}

	setup := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Modify: environment.HelloWorld{
					Runner:   &r,
					Topology: &topology,
				},
			},
		},
	}

	setup.Run()
}
