//go:build fixes
// +build fixes

package fixes

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/tester"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/frame2/walk"
	"github.com/skupperproject/skupper/test/utils/base"
)

func Test311(t *testing.T) {
	var runner = &base.ClusterTestRunnerBase{}
	var pub = runner.GetPublicContextPromise(1)
	var prv = runner.GetPrivateContextPromise(1)
	var prv2 = runner.GetPrivateContextPromise(2)
	var retryAllow10 = frame2.RetryOptions{
		Allow: 100,
	}

	var tests = frame2.TestRun{
		Name: "Test311",
		Setup: []frame2.Step{
			{
				Modify: walk.SegmentSetup{
					Namespace: prv,
				},
				// TODO Move all below to a 'setup hello world' kind of thing
			}, {
				Doc: "deploy Skupper on prv2",
				Modify: execute.SkupperInstallSimple{
					Namespace: prv2,
				},
			}, {
				Doc: "connect prv to prv2",
				Modify: execute.SkupperConnect{
					Name: "third",
					From: prv,
					To:   prv2,
				},
			}, {
				Doc: "Create frontend service",
				Modify: execute.K8SServiceCreate{
					Namespace: pub,
					Name:      "hello-world-frontend",
					Selector:  map[string]string{"app": "hello-world-frontend"},
					Labels:    map[string]string{"app": "hello-world-frontend"},
					Ports:     []int32{8080},
				},
				Validator: validate.Curl{
					Namespace: pub,
					Url:       "http://hello-world-frontend:8080",
				},
				ValidatorRetry: retryAllow10,
			}, {
				Doc: "Create backend service",
				Modify: execute.K8SServiceCreate{
					Namespace: prv,
					Name:      "hello-world-backend",
					Selector:  map[string]string{"app": "hello-world-backend"},
					Labels:    map[string]string{"app": "hello-world-backend"},
					Ports:     []int32{8080},
				},
				Validator: validate.Curl{
					Namespace: prv,
					Url:       "http://hello-world-backend:8080/api/hello",
				},
				ValidatorRetry: retryAllow10,
			},
			{
				Doc: "Annotating frontend for Skupper",
				Modify: execute.K8SServiceAnnotate{
					Namespace: pub,
					Name:      "hello-world-frontend",
					Annotations: map[string]string{
						types.ProxyQualifier:   "tcp",
						types.AddressQualifier: "hello-world-frontend",
					},
				},
				Validator: validate.Curl{
					Namespace: prv,
					Url:       "http://hello-world-frontend:8080",
				},
				ValidatorRetry: retryAllow10,
			}, {
				Doc: "Annotating backend for Skupper",
				Modify: execute.K8SServiceAnnotate{
					Namespace: prv,
					Name:      "hello-world-backend",
					Annotations: map[string]string{
						types.ProxyQualifier:   "tcp",
						types.AddressQualifier: "hello-world-backend",
					},
				},
				Validator: validate.Curl{
					Namespace: pub,
					Url:       "http://hello-world-backend:8080/api/hello",
				},
				ValidatorRetry: retryAllow10,
			},
		},
		Teardown: []frame2.Step{
			{
				Modify: walk.SegmentTeardown{
					Step: frame2.Step{Namespace: prv},
				},
			},
		},
		MainSteps: []frame2.Step{
			{
				Name: "setup-verify",
				Doc:  "are links showing good?  No retries, as previous step already checked connectivity via curl, and no changes have been made since",
				Substeps: []*frame2.Step{
					{
						Doc: "check-private-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "public",
									Active: true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
							RetryOptions:   &frame2.RetryOptions{},
						},
					}, {
						Doc: "check-public-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-hello-world-1",
									Active:          true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
							RetryOptions:   &frame2.RetryOptions{},
						},
					},
				},
			}, {
				Name: "stop-public-router",
				Doc:  "When the public router goes down, the private shows it as down",
				Substeps: []*frame2.Step{
					{
						Doc: "stop the router, generate traffic",
						Modify: execute.DeployScale{
							Namespace: *pub,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 0,
						},
						Validator: validate.Curl{
							Namespace: prv,
							Url:       "http://hello-world-frontend:8080",
						},
						ValidatorRetry: frame2.RetryOptions{
							Ignore: 2,
							Allow:  2,
						},
						ExpectError: true,
					}, {
						Name: "check-private-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "public",
									Active: false,
								},
							},
							//StrictIncoming: true,
							StrictOutgoing: true,
						},
					}, {
						Name: "check-public-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 3 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-hello-world-1",
									Active:          true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
						},
					},
				},
			}, {
				Name: "restart-public-router",
				Substeps: []*frame2.Step{
					{
						Doc: "starting the router",
						Modify: execute.DeployScale{
							Namespace: *pub,
							DeploySelector: execute.DeploySelector{
								Name: types.TransportDeploymentName,
							},
							Replicas: 1,
						},
					}, {
						Name: "check-private-outgoing",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 10 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: prv,
								},
							},
							Outgoing: []tester.CliLinkStatusOutgoing{
								{
									Name:   "public",
									Active: true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
							RetryOptions:   &retryAllow10,
						},
					}, {
						Name: "check-public-incoming",
						Modify: tester.CliLinkStatus{
							CliLinkStatus: execute.CliLinkStatus{
								Timeout: 10 * time.Second,
								CliSkupper: execute.CliSkupper{
									ClusterContext: pub,
								},
							},
							Incoming: []tester.CliLinkStatusIncoming{
								{
									SourceNamespace: "private-hello-world-1",
									Active:          true,
								},
							},
							StrictIncoming: true,
							StrictOutgoing: true,
							RetryOptions:   &retryAllow10,
						},
					},
				},
			},
		},
	}

	tests.Run(t)

}
