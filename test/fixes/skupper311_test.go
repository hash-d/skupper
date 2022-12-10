//go:build fixes
// +build fixes

package fixes

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/tester"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/frame2/walk"
	"github.com/skupperproject/skupper/test/utils/base"
)

func Test311(t *testing.T) {
	var runner = &base.ClusterTestRunnerBase{}
	var pub1 = runner.GetPublicContextPromise(1)
	var prv1 = runner.GetPrivateContextPromise(1)
	var pub2 = runner.GetPublicContextPromise(2)
	var prv2 = runner.GetPrivateContextPromise(2)
	//	var prv2 = runner.GetPrivateContextPromise(2)
	var retryAllow10 = frame2.RetryOptions{
		Allow: 120,
	}

	topologyN := topology.N{
		Name:           "test-311",
		TestRunnerBase: runner,
	}
	err := topologyN.Execute()
	if err != nil {
		t.Fatalf("failed creating topology: %v", err)
	}
	private1Ok := &frame2.Step{
		Name: "check-private1-outgoing",
		Modify: tester.CliLinkStatus{
			CliLinkStatus: execute.CliLinkStatus{
				Timeout: 3 * time.Second,
				CliSkupper: execute.CliSkupper{
					ClusterContext: prv1,
				},
			},
			Outgoing: []tester.CliLinkStatusOutgoing{
				{
					Name:   "private-test-311-1-to-public-test-311-1",
					Active: true,
				}, {
					Name:   "private-test-311-1-to-public-test-311-2",
					Active: true,
				},
			},
			StrictIncoming: true,
			StrictOutgoing: true,
			RetryOptions:   &retryAllow10,
		},
	}

	public1Ok := &frame2.Step{
		Name: "check-public1-incoming",
		Modify: tester.CliLinkStatus{
			CliLinkStatus: execute.CliLinkStatus{
				Timeout: 3 * time.Second,
				CliSkupper: execute.CliSkupper{
					ClusterContext: pub1,
				},
			},
			Incoming: []tester.CliLinkStatusIncoming{
				{
					SourceNamespace: "private-test-311-1",
					Active:          true,
				},
			},
			StrictIncoming: true,
			StrictOutgoing: true,
			RetryOptions:   &retryAllow10,
		},
	}
	private2Ok := &frame2.Step{
		Name: "check-private2-outgoing",
		Modify: tester.CliLinkStatus{
			CliLinkStatus: execute.CliLinkStatus{
				Timeout: 3 * time.Second,
				CliSkupper: execute.CliSkupper{
					ClusterContext: prv2,
				},
			},
			Outgoing: []tester.CliLinkStatusOutgoing{
				{
					Name:   "private-test-311-2-to-public-test-311-2",
					Active: true,
				},
			},
			StrictIncoming: true,
			StrictOutgoing: true,
			RetryOptions:   &retryAllow10,
		},
	}
	public2Ok := &frame2.Step{
		Name: "check-public2-incoming",
		Modify: tester.CliLinkStatus{
			CliLinkStatus: execute.CliLinkStatus{
				Timeout: 3 * time.Second,
				CliSkupper: execute.CliSkupper{
					ClusterContext: pub2,
				},
			},
			Incoming: []tester.CliLinkStatusIncoming{
				{
					SourceNamespace: "private-test-311-1",
					Active:          true,
				}, {
					SourceNamespace: "private-test-311-2",
					Active:          true,
				},
			},
			StrictIncoming: true,
			StrictOutgoing: true,
			RetryOptions:   &retryAllow10,
		},
	}

	// Ensure all links up
	var allGood = []*frame2.Step{
		private1Ok,
		public1Ok,
		private2Ok,
		public2Ok,
	}

	var tests = frame2.Phase{
		Name: "Test311",
		Doc: "Checks how routers going down impact network link status output.  " +
			"It uses an N topology (pub1 <- prv1 -> pub2 <- prv2)",
		Setup: []frame2.Step{
			{
				Modify: environment.HelloWorld{
					TopologyMap: *topologyN.Return,
				},
				// Move the ones below as an option to HelloWorld
			}, {
				Doc: "Create frontend service",
				Modify: execute.K8SServiceCreate{
					Namespace: pub1,
					Name:      "hello-world-frontend",
					Selector:  map[string]string{"app": "hello-world-frontend"},
					Labels:    map[string]string{"app": "hello-world-frontend"},
					Ports:     []int32{8080},
				},
				Validator: validate.Curl{
					Namespace: pub1,
					Url:       "http://hello-world-frontend:8080",
				},
				ValidatorRetry: retryAllow10,
			}, {
				Doc: "Create backend service",
				Modify: execute.K8SServiceCreate{
					Namespace: prv1,
					Name:      "hello-world-backend",
					Selector:  map[string]string{"app": "hello-world-backend"},
					Labels:    map[string]string{"app": "hello-world-backend"},
					Ports:     []int32{8080},
				},
				Validator: validate.Curl{
					Namespace: prv1,
					Url:       "http://hello-world-backend:8080/api/hello",
				},
				ValidatorRetry: retryAllow10,
			},
			{
				Doc: "Annotating frontend for Skupper",
				Modify: execute.K8SServiceAnnotate{
					Namespace: pub1,
					Name:      "hello-world-frontend",
					Annotations: map[string]string{
						types.ProxyQualifier:   "tcp",
						types.AddressQualifier: "hello-world-frontend",
					},
				},
				Validator: validate.Curl{
					Namespace: prv1,
					Url:       "http://hello-world-frontend:8080",
				},
				ValidatorRetry: retryAllow10,
			}, {
				Doc: "Annotating backend for Skupper",
				Modify: execute.K8SServiceAnnotate{
					Namespace: prv1,
					Name:      "hello-world-backend",
					Annotations: map[string]string{
						types.ProxyQualifier:   "tcp",
						types.AddressQualifier: "hello-world-backend",
					},
				},
				Validator: validate.Curl{
					Namespace: pub1,
					Url:       "http://hello-world-backend:8080/api/hello",
				},
				ValidatorRetry: retryAllow10,
			},
		},
		Teardown: []frame2.Step{
			{
				Modify: walk.SegmentTeardown{
					Step: frame2.Step{Namespace: prv1},
				},
			},
		},
		MainSteps: []frame2.Step{
			{
				Name:     "setup-verify",
				Doc:      "are links showing good?",
				Substeps: allGood,
			}, {
				// TODO move this to an Action
				Name: "stop-public1-router",
				Doc:  "When the public1 router goes down, the private1 shows it as down; the rest is unnafected",
				Substeps: []*frame2.Step{
					{
						Doc: "stop the public1 router, generate traffic",
						Modify: execute.DeployScale{
							Namespace: *pub1,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 0,
						},
						Validator: validate.Curl{
							Namespace: prv1,
							Url:       "http://hello-world-frontend:8080",
						},
						ValidatorRetry: frame2.RetryOptions{
							Ignore: 2,
							Allow:  2,
						},
						ExpectError: true,
					},
					private2Ok,
					public2Ok,
					{
						Name: "check-private1-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv1,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "private-test-311-1-to-public-test-311-1",
									Active: false,
								}, {
									Name:   "private-test-311-1-to-public-test-311-2",
									Active: true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						// TODO ALL DOWN
						Name: "check-public1-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub1,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-test-311-1",
									Active:          true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					},
				},
			}, {
				Name: "restart-public1-router",
				Substeps: []*frame2.Step{
					{
						Doc: "starting the router",
						Modify: execute.DeployScale{
							Namespace: *pub1,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 1,
						},
						Substeps: allGood,
					},
				},
			}, {
				// TODO move this to an Action
				Name: "stop-private1-router",
				Doc:  "When the private1 router goes down, both public show it as down; private2 is unnafected",
				Substeps: []*frame2.Step{
					{
						Doc: "stop the private1 router, generate traffic",
						Modify: execute.DeployScale{
							Namespace: *prv1,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 0,
						},
						Validator: validate.Curl{
							Namespace: pub2,
							Url:       "http://hello-world-backend:8080/api/hello",
						},
						ValidatorRetry: frame2.RetryOptions{
							Ignore: 2,
							Allow:  2,
						},
						ExpectError: true,
					},
					private2Ok,
					{
						// TODO ALL DOWN
						Name: "check-private1-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv1,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "private-test-311-1-to-public-test-311-1",
									Active: false,
								}, {
									Name:   "private-test-311-1-to-public-test-311-2",
									Active: false,
								},
							},
							//StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						Name: "check-public1-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub1,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-test-311-1",
									Active:          false,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						Name: "check-public2-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub2,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-test-311-1",
									Active:          false,
								}, {
									SourceNamespace: "private-test-311-2",
									Active:          true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					},
				},
			}, {
				Name: "restart-private1-router",
				Substeps: []*frame2.Step{
					{
						Doc: "starting the router",
						Modify: execute.DeployScale{
							Namespace: *prv1,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 1,
						},
						Substeps: allGood,
					},
				},
			}, {
				// TODO move this to an Action
				Name: "stop-public2-router",
				Doc:  "When the public2 router goes down, both private show it as down; public1 is unnafected",
				Substeps: []*frame2.Step{
					{
						Doc: "stop the public2 router, generate traffic",
						Modify: execute.DeployScale{
							Namespace: *pub2,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 0,
						},
						Validator: validate.Curl{
							Namespace: prv2,
							Url:       "http://hello-world-frontend:8080",
						},
						ValidatorRetry: frame2.RetryOptions{
							Ignore: 2,
							Allow:  2,
						},
						ExpectError: true,
					},
					public1Ok,
					{
						Name: "check-private1-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv1,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "private-test-311-1-to-public-test-311-1",
									Active: true,
								}, {
									Name:   "private-test-311-1-to-public-test-311-2",
									Active: false,
								},
							},
							//StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						Name: "check-private2-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv2,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "private-test-311-2-to-public-test-311-2",
									Active: false,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						// TODO DOWN
						Name: "check-public2-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub2,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-test-311-1",
									Active:          false,
								},
								{
									SourceNamespace: "private-test-311-2",
									Active:          false,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					},
				},
			}, {
				Name: "restart-public2-router",
				Substeps: []*frame2.Step{
					{
						Doc: "starting the router",
						Modify: execute.DeployScale{
							Namespace: *pub2,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 1,
						},
						Substeps: allGood,
					},
				},
			}, {

				// TODO move this to an Action
				Name: "stop-private2-router",
				Doc:  "When the private2 router goes down, public2 shows one of its links down; the rest is unnafected",
				Substeps: []*frame2.Step{
					{
						Doc: "stop the private2 router, generate traffic",
						Modify: execute.DeployScale{
							Namespace: *prv2,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 0,
						},
						Validator: validate.Curl{
							Namespace: pub2,
							Url:       "http://hello-world-backend:8080/api/hello",
						},
						ValidatorRetry: frame2.RetryOptions{
							Ignore: 2,
							Allow:  2,
						},
						// No error, as private1 is connected to both public namespaces,
						// so the backend is still available
					},
					private1Ok,
					public1Ok,
					{
						// TODO all down
						Name: "check-private2-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv2,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "private-test-311-2-to-public-test-311-2",
									Active: false,
								},
							},
							//StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						Name: "check-public2-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub2,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-test-311-1",
									Active:          true,
								}, {
									SourceNamespace: "private-test-311-2",
									Active:          false,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					},
				},
			}, {
				Name: "restart-private2-router",
				Substeps: []*frame2.Step{
					{
						Doc: "starting the router",
						Modify: execute.DeployScale{
							Namespace: *prv2,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 1,
						},
						Substeps: allGood,
					},
				},
			},
		},
	}

	tests.RunT(t)

}
