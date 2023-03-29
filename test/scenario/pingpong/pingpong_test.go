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
// TODO
// - remove service first
// - remove link first
// - skupper delete, direct
// - or remove the target deployment
//
// By default, use a different one each time, but allow
// for selecting a single one
package pingpong

import (
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

// Installs HelloWorld front and backend on on the left branch of a V-shaped topology, and then
// migrates it to the right branch and back
func TestPingPong(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	defer r.Report()

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
				Doc: "Setup a HelloWorld environment",
				Modify: environment.HelloWorld{
					Runner:        &r,
					Topology:      &topologyV,
					AutoTearDown:  true,
					SkupperExpose: true,
				},
			},
		},
	}
	assert.Assert(t, setup.Run())

	var topo topology.TwoBranched = topologyV.(topology.TwoBranched)
	vertex, err := topo.GetVertex()
	assert.Assert(t, err)
	rightFront, err := topo.GetRight(topology.Public, 1)
	assert.Assert(t, err)
	leftBack, err := topo.GetLeft(topology.Private, 1)
	assert.Assert(t, err)
	leftFront, err := topo.GetLeft(topology.Public, 1)
	assert.Assert(t, err)
	rightBack, err := topo.GetRight(topology.Private, 1)
	assert.Assert(t, err)

	monitorPhase := frame2.Phase{
		Runner: &r,
		// This is Mainsteps, not setup, to ensure that the monitors are installed
		// even if the setup step was skipped.  It's also out of the loop, so we
		// install it only once
		MainSteps: []frame2.Step{
			{
				// Our validations will run from the vertex node; before we
				// start monitoring, let's make sure it looks good
				Doc: "Validate Hello World deployment from vertex",
				Validator: &deploy.HelloWorldValidate{
					Namespace: vertex,
					Runner:    &r,
				},
				ValidatorRetry: frame2.RetryOptions{
					Allow:  60,
					Ignore: 5,
					Ensure: 5,
				},
			}, {

				Doc: "Installing hello-world monitors",
				Modify: &frame2.DefaultMonitor{
					Validators: map[string]frame2.Validator{
						"hello-world": &deploy.HelloWorldValidate{
							Runner:    &r,
							Namespace: vertex,
						},
					},
				},
			},
		},
	}
	assert.Assert(t, monitorPhase.Run())

	deltas := []time.Duration{}

	for {
		startTime := time.Now()
		main := frame2.Phase{
			Runner: &r,
			MainSteps: []frame2.Step{
				{
					Modify: &execute.CliSkupper{
						Args:           []string{"network", "status"},
						ClusterContext: vertex,
						Cmd: execute.Cmd{
							ForceOutput: true,
						},
					},
				}, {
					Name: "Move to right",
					Modify: &MoveToRight{
						Runner:     &r,
						Topology:   topologyV.(topology.TwoBranched),
						LeftFront:  leftFront,
						LeftBack:   leftBack,
						RightFront: rightFront,
						RightBack:  rightBack,
						Vertex:     vertex,
					},
				}, {
					Modify: &execute.CliSkupper{
						Args:           []string{"network", "status"},
						ClusterContext: vertex,
						Cmd: execute.Cmd{
							ForceOutput: true,
						},
					},
				}, {
					Name: "Move to left",
					Modify: &MoveToLeft{
						Runner:     &r,
						Topology:   topologyV.(topology.TwoBranched),
						LeftFront:  leftFront,
						LeftBack:   leftBack,
						RightFront: rightFront,
						RightBack:  rightBack,
						Vertex:     vertex,
					},
				},
			},
		}
		assert.Assert(t, main.Run())

		// Move all the log below to a new Executor: GreedyRepeatedTester
		endTime := time.Now()

		delta := endTime.Sub(startTime)
		deltas = append(deltas, delta)

		testDeadline, ok := t.Deadline()
		if !ok {
			// No deadline, and we do not want to loop forever
			break
		}
		var maxTime time.Duration
		var totalDuration time.Duration

		for _, d := range deltas {
			totalDuration += d
			if d > maxTime {
				maxTime = d
			}
		}

		if testDeadline.Sub(time.Now()) < maxTime*2 {
			log.Printf("Finishing Pingpong test after %d run(s)", len(deltas))
			log.Printf(
				"The average pingpong was %v; max was %v",
				totalDuration/time.Duration(len(deltas)),
				maxTime,
			)
			return
		}

	}
}

type MoveToRight struct {
	Runner     *frame2.Run
	Topology   topology.TwoBranched
	Vertex     *base.ClusterContext
	LeftFront  *base.ClusterContext
	LeftBack   *base.ClusterContext
	RightFront *base.ClusterContext
	RightBack  *base.ClusterContext
}

// TODO: can this be made more generic, instead?
func (m *MoveToRight) Execute() error {

	log.Printf("LF: %+v\nLB: %+v\nRF: %+v\nRB: %+v\nVX: %+v\n", m.LeftFront, m.LeftBack, m.RightFront, m.RightBack, m.Vertex)
	validateHW := deploy.HelloWorldValidate{
		Runner:    m.Runner,
		Namespace: m.Vertex,
	}
	validateOpts := frame2.RetryOptions{
		Allow:  5,
		Ignore: 5,
		Ensure: 5,
	}

	p := frame2.Phase{
		Runner: m.Runner,
		Doc:    "Move Hello World from left to right",
		MainSteps: []frame2.Step{
			{
				Doc: "Move frontend from left to right",
				Modify: &composite.Migrate{
					Runner:     m.Runner,
					From:       m.LeftFront,
					To:         m.RightFront,
					LinkTo:     []*base.ClusterContext{},
					LinkFrom:   []*base.ClusterContext{m.LeftBack, m.Vertex},
					UnlinkFrom: []*base.ClusterContext{m.Vertex},
					DeploySteps: []frame2.Step{
						{
							Doc: "Deploy new HelloWorld Frontend",
							Modify: &deploy.HelloWorldFrontend{
								Runner:        m.Runner,
								Target:        m.RightFront,
								SkupperExpose: true,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
					UndeploySteps: []frame2.Step{
						{
							Doc: "Remove the application from the old frontend namespace",
							Modify: &execute.K8SUndeploy{
								Name:      "hello-world-frontend",
								Namespace: m.LeftFront,
								Wait:      2 * time.Minute,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
				},
			}, {
				Doc: "Move backend from left to right",
				Modify: &composite.Migrate{
					From:     m.LeftBack,
					To:       m.RightBack,
					LinkTo:   []*base.ClusterContext{m.RightFront},
					LinkFrom: []*base.ClusterContext{},
					DeploySteps: []frame2.Step{
						{
							Doc: "Deploy new HelloWorld Backend",
							Modify: &deploy.HelloWorldBackend{
								Target: m.RightBack,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
					UndeploySteps: []frame2.Step{
						{
							Doc: "Remove the application from the old backend namespace",
							Modify: &execute.K8SUndeploy{
								Name:      "hello-world-backend",
								Namespace: m.LeftBack,
								Wait:      2 * time.Minute,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
				},
			},
		},
	}

	p.Run()

	return nil
}

type MoveToLeft struct {
	Runner     *frame2.Run
	Topology   topology.TwoBranched
	Vertex     *base.ClusterContext
	LeftFront  *base.ClusterContext
	LeftBack   *base.ClusterContext
	RightFront *base.ClusterContext
	RightBack  *base.ClusterContext
}

// TODO: can this be made more generic, instead?
func (m *MoveToLeft) Execute() error {

	log.Printf("LF: %+v\nLB: %+v\nRF: %+v\nRB: %+v\nVX: %+v\n", m.LeftFront, m.LeftBack, m.RightFront, m.RightBack, m.Vertex)
	validateHW := deploy.HelloWorldValidate{
		Runner:    m.Runner,
		Namespace: m.Vertex,
	}
	validateOpts := frame2.RetryOptions{
		Allow:  5,
		Ignore: 5,
		Ensure: 5,
	}

	p := frame2.Phase{
		Runner: m.Runner,
		Doc:    "Move Hello World from right to left",
		MainSteps: []frame2.Step{
			{
				Doc: "Move frontend from right to left",
				Modify: &composite.Migrate{
					Runner:     m.Runner,
					From:       m.RightFront,
					To:         m.LeftFront,
					LinkTo:     []*base.ClusterContext{},
					LinkFrom:   []*base.ClusterContext{m.RightBack, m.Vertex},
					UnlinkFrom: []*base.ClusterContext{m.Vertex},
					DeploySteps: []frame2.Step{
						{
							Doc: "Deploy new HelloWorld Frontend",
							Modify: &deploy.HelloWorldFrontend{
								Runner:        m.Runner,
								Target:        m.LeftFront,
								SkupperExpose: true,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
					UndeploySteps: []frame2.Step{
						{
							Doc: "Remove the application from the old frontend namespace",
							Modify: &execute.K8SUndeploy{
								Name:      "hello-world-frontend",
								Namespace: m.RightFront,
								Wait:      2 * time.Minute,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
				},
			}, {
				Doc: "Move backend from right to left",
				Modify: &composite.Migrate{
					From:     m.RightBack,
					To:       m.LeftBack,
					LinkTo:   []*base.ClusterContext{m.LeftFront},
					LinkFrom: []*base.ClusterContext{},
					DeploySteps: []frame2.Step{
						{
							Doc: "Deploy new HelloWorld Backend",
							Modify: &deploy.HelloWorldBackend{
								Runner: m.Runner,
								Target: m.LeftBack,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
					UndeploySteps: []frame2.Step{
						{
							Doc: "Remove the application from the old backend namespace",
							Modify: &execute.K8SUndeploy{
								Name:      "hello-world-backend",
								Namespace: m.RightBack,
								Wait:      2 * time.Minute,
							},
							Validator:      &validateHW,
							ValidatorRetry: validateOpts,
						},
					},
				},
			},
		},
	}

	p.Run()

	return nil
}
