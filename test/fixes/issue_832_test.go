package fixes

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
	v1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestIssue832(t *testing.T) {
	runner := &frame2.Run{T: t}

	baseRunner := base.ClusterTestRunnerBase{}

	// These two phases should be combined into a single one,
	// with a composite Executor
	topoMap := &topology.Simplest{
		Name:           "issue-832",
		TestRunnerBase: &baseRunner,
	}

	phase1 := frame2.Phase{
		Runner: runner,
		Setup: []frame2.Step{
			{
				Doc:    "Create the topology map",
				Modify: topoMap,
			},
		},
	}
	phase1.Run()

	topo := &topology.Topology{
		Runner:       runner,
		TopologyMap:  topoMap.Return,
		AutoTearDown: true,
	}

	// TODO: change this by something that's returned from topology.Topology
	//var pubPromise *base.ClusterContextPromise
	var prvPromise *base.ClusterContextPromise

	//	var pub *base.ClusterContext
	var prv *base.ClusterContext

	phase2 := frame2.Phase{
		Runner: runner,
		Setup: []frame2.Step{
			{
				Doc:    "Create the actual topology",
				Modify: topo,
			}, {
				Doc: "Get the Promises",
				Modify: execute.Function{
					Fn: func() error {
						//pubPromise = baseRunner.GetContextPromise(false, 1)
						prvPromise = baseRunner.GetContextPromise(true, 1)
						return nil
					},
				},
			}, {
				Doc: "Get the prv reference",
				Modify: execute.Function{
					Fn: func() error {
						var err error
						prv, err = prvPromise.Satisfy()
						return err
					},
				},
			}, {
				Doc: "Get the pub reference",
				Modify: execute.Function{
					Fn: func() error {
						var err error
						//pub, err = pubPromise.Satisfy()
						return err
					},
				},
			},
		},
	}

	phase2.Run()

	labels := map[string]string{
		"app": "backend",
	}

	phase3 := frame2.Phase{
		Runner: runner,
		Setup: []frame2.Step{
			{
				Doc: "Deploy an nginx server, with a readiness proble delayed to start only after 60 seconds",
				Modify: &execute.K8SDeployment{
					Namespace: prvPromise,
					Deployment: &v1.Deployment{
						ObjectMeta: v13.ObjectMeta{
							Name:      "backend",
							Namespace: prv.Namespace,
							Labels:    labels,
						},
						Spec: v1.DeploymentSpec{
							Selector: &v13.LabelSelector{
								MatchLabels: labels,
							},
							Template: core.PodTemplateSpec{
								ObjectMeta: v13.ObjectMeta{
									Labels: labels,
								},
								Spec: core.PodSpec{
									Containers: []core.Container{
										{
											Name:            "backend",
											Image:           "nginx",
											ImagePullPolicy: core.PullIfNotPresent,
											// Command:         []string{"sh", "-c"},
											// Args:            []string{""},
											Ports: []core.ContainerPort{
												{
													HostPort:      8080,
													ContainerPort: 80,
												},
											},
											ReadinessProbe: &core.Probe{
												InitialDelaySeconds: 60,
												Handler: core.Handler{
													HTTPGet: &core.HTTPGetAction{
														Path: "/",
														Port: intstr.FromInt(80),
													},
												},
											},
										},
									},
									RestartPolicy: core.RestartPolicyAlways,
								},
							},
						},
					},
				},
			},
		},
	}

	phase3.Run()

	phase4 := frame2.Phase{
		Runner: runner,
		Setup: []frame2.Step{
			{
				Doc: "wait and see",
				Modify: execute.Function{
					Fn: func() error {
						time.Sleep(5 * time.Minute)
						return nil
					},
				},
			},
		},
	}
	phase4.Run()

}
