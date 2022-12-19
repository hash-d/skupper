package fixes

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/validates"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestIssue832(t *testing.T) {
	runner := &frame2.Run{
		T: t,
		Doc: `
			Kubernetes allows to control when the DNS entry of a pod participating in a headless service gets published: at creation or only when ready.

			This test checks whether such services exposed by Skupper accept and comply with such configuration.
			`,
	}

	// Test cases:
	//
	// ~~Create stateful set without service; expose via skupper~~ This is not possible; statefulset creation works, but skupper complains
	// Create k8s service, statefulset; expose service via skupper
	// Create skupper service, then stateful set, bind
	// Same operations with annotations
	// Headless services pointing to Deployments instead of Statefulsets?

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
	var pubPromise *base.ClusterContextPromise
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
						pubPromise = baseRunner.GetContextPromise(false, 1)
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

	for _, target := range []*base.ClusterContextPromise{prvPromise, pubPromise} {
		// TODO move this to a Executor "DeployDNStools
		deployDnsTooling := frame2.Phase{
			Runner: runner,
			MainSteps: []frame2.Step{
				{
					Modify: &execute.K8SDeploymentOpts{
						Name:      "dnsutils",
						Namespace: target,
						DeploymentOpts: k8s.DeploymentOpts{
							Image:         "registry.k8s.io/e2e-test-images/jessie-dnsutils:1.7",
							Labels:        map[string]string{"app": "dnsutil"},
							Command:       []string{"sleep"},
							Args:          []string{"infinity"},
							RestartPolicy: core.RestartPolicyAlways,
						},
						Wait: 2 * time.Minute,
					},
				},
			},
		}
		deployDnsTooling.Run()
	}

	var baseSf apps.StatefulSet

	labels := map[string]string{
		"app": "backend",
	}

	prepareBaseSf := frame2.Phase{
		Runner: runner,
		Doc:    "Create a base statefulset, that's going to be modified for each test, it deploys a hello-world-backend with a ReadinessProbe configured to start after 60 seconds",
		Setup: []frame2.Step{
			{
				Modify: execute.Function{
					Fn: func() error {
						baseSf = apps.StatefulSet{
							ObjectMeta: meta.ObjectMeta{
								Name:      "backend",
								Namespace: prv.Namespace,
								Labels:    labels,
							},
							Spec: apps.StatefulSetSpec{
								// More than one replica on a simpler test/dev environment (such as Minikube with only one
								// node) may cause issues.
								// Replicas: &replicas,
								Selector: &meta.LabelSelector{
									MatchLabels: labels,
								},
								ServiceName: "backend",
								Template: core.PodTemplateSpec{
									ObjectMeta: meta.ObjectMeta{
										Labels: labels,
									},
									Spec: core.PodSpec{
										Containers: []core.Container{
											{
												Name:            "backend",
												Image:           "quay.io/skupper/hello-world-backend",
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
															Path: "/api/hello",
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
						}
						return nil
					},
				},
			},
		},
	}
	prepareBaseSf.Run()

	bindSkupperBased := frame2.Phase{
		Runner: runner,
		Name:   "bind-skupper-based-stateful",
		Doc:    "Skupper service basing a statefulset; the statefulset is bound to the service",
		Setup: []frame2.Step{
			{
				Doc: "Create a Skupper service for the statefulset",
				Modify: execute.SkupperServiceCreate{
					Runner:       runner,
					Namespace:    prvPromise,
					Name:         "backend",
					Port:         []string{"8080"},
					AutoTeardown: true,
				},
			}, {
				Doc: "Deploy a hello-world-backend server, with a readiness probe delayed to start only after 60 seconds",
				Modify: &execute.K8SStatefulSet{
					Namespace:    prvPromise,
					AutoTeardown: true,
					StatefulSet:  &baseSf,
				},
			}, {
				Doc: "Bind the stateful set; expect the name to not be available for at least 30s",
				Modify: execute.SkupperServiceBind{
					Runner:                 runner,
					Namespace:              prvPromise,
					Name:                   "backend",
					TargetType:             "statefulset",
					TargetName:             "backend",
					PublishNotReadyAddress: true,
					AutoTeardown:           true,
				},
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend",
					},
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   30,
				},
				//ExpectError: true,
			}, {
				Doc: "After the initial 30s from the previous step, give at most 1m for it to be up",
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend",
					},
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   3,
					Allow:    60,
				},
			},
		},
	}
	bindSkupperBased.Run()

	exposePlainSf := frame2.Phase{
		Runner: runner,
		Name:   "expose-plain-statefulset",
		Doc:    "expose a statefulset based by a headless k8s service",
		Setup: []frame2.Step{
			{
				Doc: "Create a k8s service for the statefulset",
				Modify: execute.K8SServiceCreate{
					// TODO add Runner to this guy
					// Runner:    runner,
					Namespace: prvPromise,
					Name:      "backend",
					Ports:     []int32{8080},
					Labels:    labels,
					Selector:  labels,
					ClusterIP: "None",
					Type:      core.ServiceTypeClusterIP,
					//AutoTeardown: true,
				},
			}, {
				Doc: "Deploy a hello-world-backend server, with a readiness probe delayed to start only after 60 seconds",
				Modify: &execute.K8SStatefulSet{
					Namespace:    prvPromise,
					AutoTeardown: true,
					StatefulSet:  &baseSf,
				},
			}, {
				Doc: "Expose the stateful set",
				Modify: execute.SkupperExpose{
					Runner:    runner,
					Namespace: prvPromise,
					Name:      "backend",
					Type:      "statefulset",
					Headless:  true,
					//PublishNotReadyAddress: true,
					Address: "backend",
					//AutoTeardown:           true,
				},
			}, {
				Doc: "Bind the stateful set; expect the name to not be available for at least 30s",
				Modify: execute.SkupperServiceBind{
					Runner:                 runner,
					Namespace:              prvPromise,
					Name:                   "backend",
					TargetType:             "statefulset",
					TargetName:             "backend",
					PublishNotReadyAddress: true,
					AutoTeardown:           true,
				},
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend",
					},
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   30,
				},
				//ExpectError: true,
			}, {
				Doc: "After the initial 30s from the previous step, give at most 1m for it to be up",
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend",
					},
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   3,
					Allow:    60,
				},
			},
		},
	}

	exposePlainSf.Run()

	phase4 := frame2.Phase{
		Runner: runner,
		Setup: []frame2.Step{
			{
				Doc: "wait and see",
				Modify: execute.Function{
					Fn: func() error {
						time.Sleep(1 * time.Minute)
						return nil
					},
				},
			},
		},
	}
	phase4.Run()

}
