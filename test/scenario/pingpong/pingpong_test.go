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
// - or remove the target deployment
//
// By default, use a different one each time, but allow
// for selecting a single one
package pingpong

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/composite"
	"github.com/skupperproject/skupper/test/frame2/deploy"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

var runner = &base.ClusterTestRunnerBase{}

func TestPingPong(t *testing.T) {
	r := frame2.Run{
		T: t,
	}

	var topologyV topology.Basic
	topologyV = &topologies.V{
		Name:           "pingpong",
		TestRunnerBase: runner,
		EmptyRight:     true,
	}

	setup := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Modify: environment.HelloWorld{
					Runner:       &r,
					Topology:     &topologyV,
					AutoTearDown: true,
				},
			},
		},
	}

	assert.Assert(t, setup.Run())

	vertex, err := topologyV.Get(topology.Public, 3)
	assert.Assert(t, err)

	main := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Modify: &execute.CliSkupper{
					Args:           []string{"network", "status"},
					ClusterContext: vertex.GetPromise(),
				},
			}, {
				Modify: &MoveToRight{
					Runner:   &r,
					Topology: topologyV.(topology.TwoBranched),
				},
			},
		},
	}

	assert.Assert(t, main.Run())
}

type MoveToRight struct {
	Runner   *frame2.Run
	Topology topology.TwoBranched
}

// TODO: can this be made more generic, instead?
func (m *MoveToRight) Execute() error {

	rightFront, err := m.Topology.GetRight(topology.Public, 1)
	if err != nil {
		return fmt.Errorf("MoveToRight: failed to get right frontend namespace: %w", err)
	}
	leftBack, err := m.Topology.GetLeft(topology.Private, 1)
	if err != nil {
		return fmt.Errorf("MoveToRight: failed to get left backend namespace: %w", err)
	}
	vertex, err := m.Topology.GetVertex()
	if err != nil {
		return fmt.Errorf("MoveToRight: failed to get vertex: %w", err)
	}
	leftFront, err := m.Topology.GetLeft(topology.Public, 1)
	if err != nil {
		return fmt.Errorf("MoveToRight: failed to get left frontend namespace: %w", err)
	}
	rightBack, err := m.Topology.GetRight(topology.Private, 1)
	if err != nil {
		return fmt.Errorf("MoveToRight: failed to get right backend namespace: %w", err)
	}

	log.Printf("LF: %+v\nLB: %+v\nRF: %+v\nRB: %+v\nVX: %+v\n", leftFront, leftBack, rightFront, rightBack, vertex)

	p := frame2.Phase{
		Runner: m.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Move frontend from left to right",
				Modify: &composite.Migrate{
					From:     leftFront,
					To:       rightFront,
					LinkTo:   []*base.ClusterContext{},
					LinkFrom: []*base.ClusterContext{leftBack, vertex},
					DeploySteps: []frame2.Step{
						{
							Doc: "Deploy new HelloWorld Frontend",
							Modify: &deploy.HelloWorldFrontend{
								Target: rightFront.GetPromise(),
							},
						},
					},
					UndeploySteps: []frame2.Step{
						{
							Doc: "Remove the application from the old frontend namespace",
							Modify: &execute.K8SUndeploy{
								Name:      "hello-world-frontend",
								Namespace: leftFront.GetPromise(),
								Wait:      2 * time.Minute,
							},
						},
					},
				},
			}, {
				Doc: "Move backend from left to right",
				Modify: &composite.Migrate{
					From:     leftBack,
					To:       rightBack,
					LinkTo:   []*base.ClusterContext{rightFront},
					LinkFrom: []*base.ClusterContext{},
					DeploySteps: []frame2.Step{
						{
							Doc: "Deploy new HelloWorld Backend",
							Modify: &deploy.HelloWorldBackend{
								Target: rightBack.GetPromise(),
							},
						},
					},
					UndeploySteps: []frame2.Step{
						{
							Doc: "Remove the application from the old backend namespace",
							Modify: &execute.K8SUndeploy{
								Name:      "hello-world-backend",
								Namespace: leftBack.GetPromise(),
								Wait:      2 * time.Minute,
							},
						},
					},
				},
			},
		},
	}

	p.Run()

	return nil
}
