package fixes

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/validate"
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
	// Create k8s service, statefulset; expose service via skupper
	// ~~Create stateful set without service; expose via skupper~~ This is not possible; statefulset creation works, but skupper complains
	// ~~Create skupper service, then stateful set, bind~~ Invalid test case; skupper service is not headless
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

	// Install dnsutils so we can check the published names
	for _, target := range []*base.ClusterContextPromise{prvPromise, pubPromise} {
		// TODO move this to a Executor "DeployDNStools
		//      also, good candidate for parallelism
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
						Wait:   2 * time.Minute,
						Runner: runner,
					},
				},
			},
		}
		deployDnsTooling.Run()
	}

	// The basic statefulset to be created on the tests.  If different statefulset configurations need created,
	// they can be based on copies of this.
	var baseSf apps.StatefulSet

	labels := map[string]string{
		"app": "backend",
	}

	prepareBaseSf := frame2.Phase{
		Runner: runner,
		Doc: `
			Create a base statefulset, that's going to be modified for each test; it deploys a hello-world-backend with a ReadinessProbe configured to start after 90 seconds.

			This delay on the probe allows us to inspect whether the dns name is being exposed before the stateful set is ready.
			`,
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
														ContainerPort: 8080,
													},
												},
												ReadinessProbe: &core.Probe{
													InitialDelaySeconds: 90,
													Handler: core.Handler{
														HTTPGet: &core.HTTPGetAction{
															Path: "/api/hello",
															Port: intstr.FromInt(8080),
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

	control := frame2.Phase{
		Runner: runner,
		Name:   "control-group",
		Doc: `This is the 'control group'; a normal service backing the statefulset, which is exposed by
			Skupper without publishNotReadyAdress; the name is supposed to not appear before the statefulset
			reports as ready`,
		Setup: []frame2.Step{
			{
				Doc: "Create a headless k8s service for the statefulset",
				Modify: execute.K8SServiceCreate{
					Namespace:    prvPromise,
					Name:         "backend",
					Ports:        []int32{8080},
					Labels:       labels,
					Selector:     labels,
					ClusterIP:    "None",
					Type:         core.ServiceTypeClusterIP,
					AutoTeardown: true,
					Wait:         time.Minute,
				},
			}, {
				Doc: "Deploy a hello-world-backend server, with a readiness probe delayed to start only after 90 seconds",
				Modify: &execute.K8SStatefulSet{
					Namespace:    prvPromise,
					AutoTeardown: true,
					StatefulSet:  &baseSf,
				},
			}, {
				Doc: "Expose the stateful set; expect the name to not be available for at least 60s",
				Modify: execute.SkupperExpose{
					Runner:                 runner,
					Namespace:              prvPromise,
					Name:                   "backend",
					Type:                   "statefulset",
					Headless:               true,
					PublishNotReadyAddress: false,
					AutoTeardown:           true,
				},
				Validators: []frame2.Validator{
					// This is failing; described as Finding 2 / Problem B on SKUPPER-310
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend-0.backend",
					},
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend-0.backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   60, // For at least the initial 60s, the names should _not_ be published.
				},
				ExpectError: true,
			}, {
				Doc: "After the initial 60s from the previous step, give at most 2m for the name to appear",
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend-0.backend",
					},
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend-0.backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   3,
					Allow:    120,
				},
			}, {
				Doc: "Ensures that the backend is actually available, by accessing it via Curl",
				Validators: []frame2.Validator{
					validate.Curl{
						Namespace: prvPromise,
						Url:       "http://backend-0.backend:8080/api/hello",
					},
					validate.Curl{
						Namespace: pubPromise,
						Url:       "http://backend-0.backend:8080/api/hello",
					},
				},
			},
		},
	}
	control.Run()

	/*
		// This is not a valid scenario.  According to https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/:
		//
		// "StatefulSets currently require a Headless Service to be responsible
		// for the network identity of the Pods. You are responsible for
		// creating this Service"
		//
		// The service that Skupper creates with skupper service create is not
		// a headless service
		//
		bindSkupperBased := frame2.Phase{
			Runner: runner,
			Name:   "bind-skupper-based-stateful",
			Doc:    "Skupper service basing a statefulset; the statefulset gets bound to the service",
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
					Doc: "Deploy a hello-world-backend server, with a readiness probe delayed to start only after 90 seconds",
					Modify: &execute.K8SStatefulSet{
						Namespace:    prvPromise,
						AutoTeardown: true,
						StatefulSet:  &baseSf,
					},
				}, {
					Doc: "Bind the stateful set; expect the name to be available before the actual pod is ready",
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
							Name:      "backend-0.backend",
						},
						validates.Nslookup{
							Namespace: pubPromise,
							Name:      "backend-0.backend",
						},
					},
					ValidatorRetry: frame2.RetryOptions{
						Interval: time.Second,
						Ensure:   30,
						Allow:    10,
						Retries:  100,
					},
				},
			},
		}
		bindSkupperBased.Run()
	*/

	exposePlainSf := frame2.Phase{
		Runner: runner,
		Name:   "expose-plain-statefulset",
		Doc:    "expose a statefulset based by a headless k8s service",
		Setup: []frame2.Step{
			{
				Doc: "Create a headless k8s service for the statefulset",
				Modify: execute.K8SServiceCreate{
					// TODO add Runner to this guy
					// Runner:    runner,
					Namespace:                prvPromise,
					Name:                     "backend",
					Ports:                    []int32{8080},
					Labels:                   labels,
					Selector:                 labels,
					ClusterIP:                "None",
					Type:                     core.ServiceTypeClusterIP,
					PublishNotReadyAddresses: true,
					AutoTeardown:             true,
				},
			}, {
				Doc: "Deploy a hello-world-backend server, with a readiness probe delayed to start only after 90 seconds",
				Modify: &execute.K8SStatefulSet{
					Namespace:    prvPromise,
					StatefulSet:  &baseSf,
					AutoTeardown: true,
				},
			}, {
				Doc: "Expose the stateful set via Skupper; expect the names to be almost immediatelly available",
				Modify: execute.SkupperExpose{
					Runner:                 runner,
					Namespace:              prvPromise,
					Name:                   "backend",
					Type:                   "statefulset",
					Headless:               true,
					PublishNotReadyAddress: true,
					Address:                "backend",
					AutoTeardown:           true,
				},
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "backend-0.backend",
					},
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "backend-0.backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   3,
					// The name should be available fairly soon, and way before the
					// probe delay we set
					Allow: 30,
				},
			}, {
				Doc: "Ensures that the backend is actually available, by accessing it via Curl",
				Validators: []frame2.Validator{
					validate.Curl{
						Namespace: prvPromise,
						Url:       "http://backend-0.backend:8080/api/hello",
					},
					validate.Curl{
						Namespace: pubPromise,
						Url:       "http://backend-0.backend:8080/api/hello",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Ensure: 3,
					// The name is expected to be available before the actual backend.
					// For that reason, this may fail several times, until the backend
					// is up.
					Allow: 120,
				},
			},
		},
	}
	exposePlainSf.Run()

	// This is failing, described as Finding 3 / Problem C on SKUPPER-310
	bindPlainSf := frame2.Phase{
		Runner: runner,
		Name:   "bind-plain-statefulset",
		Doc:    "bind a statefulset based by a headless k8s service, to a new skupper service address",
		Setup: []frame2.Step{
			{
				Doc: "Create a headless k8s service for the statefulset",
				Modify: execute.K8SServiceCreate{
					// TODO add Runner to this guy
					// Runner:    runner,
					Namespace:                prvPromise,
					Name:                     "backend",
					Ports:                    []int32{8080},
					Labels:                   labels,
					Selector:                 labels,
					ClusterIP:                "None",
					Type:                     core.ServiceTypeClusterIP,
					PublishNotReadyAddresses: true,
					AutoTeardown:             true,
				},
			}, {
				Doc: "Deploy a hello-world-backend server, with a readiness probe delayed to start only after 90 seconds",
				Modify: &execute.K8SStatefulSet{
					Namespace:    prvPromise,
					StatefulSet:  &baseSf,
					AutoTeardown: true,
				},
			}, {
				Doc: "Create a new skupper service, named skupper-backend",
				Modify: &execute.SkupperServiceCreate{
					Namespace:    prvPromise,
					Name:         "skupper-backend",
					Port:         []string{"8080"},
					Runner:       runner,
					AutoTeardown: true,
				},
			}, {
				Doc: "Bind the stateful set via Skupper; expect the names to be almost immediatelly available",
				Modify: execute.SkupperServiceBind{
					Runner:                 runner,
					Namespace:              prvPromise,
					Name:                   "skupper-backend",
					TargetType:             "statefulset",
					TargetName:             "backend",
					PublishNotReadyAddress: true,
					AutoTeardown:           true,
				},
				Validators: []frame2.Validator{
					validates.Nslookup{
						Namespace: pubPromise,
						Name:      "skupper-backend-0.skupper-backend",
					},
					validates.Nslookup{
						Namespace: prvPromise,
						Name:      "skupper-backend-0.skupper-backend",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Interval: time.Second,
					Ensure:   3,
					// The name should be available fairly soon, and way before the
					// probe delay we set
					Allow: 30,
				},
			}, {
				Doc: "Ensures that the backend is actually available, by accessing it via Curl",
				Validators: []frame2.Validator{
					validate.Curl{
						Namespace: prvPromise,
						Url:       "http://skupper-backend-0.skupper-backend:8080/api/hello",
					},
					validate.Curl{
						Namespace: pubPromise,
						Url:       "http://skupper-backend-0.skupper-backend:8080/api/hello",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Ensure: 3,
					// The name is expected to be available before the actual backend.
					// For that reason, this may fail several times, until the backend
					// is up.
					Allow: 120,
				},
			},
		},
	}
	bindPlainSf.Run()

}
