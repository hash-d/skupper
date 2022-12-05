package fixes

import (
	"testing"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/frame2/walk"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/tools"
	"gotest.tools/assert"
)

func TestSkupper314(t *testing.T) {
	assert.Assert(t, tests.Run(t))
}

var runner = &base.ClusterTestRunnerBase{}

var pub = runner.GetPublicContextPromise(1)
var prv = runner.GetPrivateContextPromise(1)

var pubBackDoesntWork = frame2.Step{
	Doc: "Validate backend from pub doesnt work",
	Validator: validate.Curl{
		Namespace:   pub,
		Url:         "http://hello-world-backend-k8s-service:8080/api/hello",
		CurlOptions: tools.CurlOpts{Timeout: 10},
	},
	ValidatorRetry: frame2.RetryOptions{
		Allow: 10,
	},
	ExpectError: true,
}

var prvFrontDoesntWork = frame2.Step{
	Doc: "Validate frontend from prv doesnt work",
	Validator: validate.Curl{
		Namespace:   prv,
		Url:         "http://hello-world-frontend-k8s-service:8080",
		CurlOptions: tools.CurlOpts{Timeout: 10},
	},
	ValidatorRetry: frame2.RetryOptions{
		Allow: 10,
	},
	ExpectError: true,
}

var pubBackWorks = frame2.Step{
	Doc: "Validate backend from pub",
	Validator: validate.Curl{
		Namespace:   pub,
		Url:         "http://hello-world-backend-k8s-service:8080/api/hello",
		CurlOptions: tools.CurlOpts{Timeout: 10},
	},
	ValidatorRetry: frame2.RetryOptions{
		Allow: 10,
	},
}

var pubFrontWorks = frame2.Step{
	Doc: "Validate frontend from pub",
	Validator: validate.Curl{
		Namespace:   pub,
		Url:         "http://hello-world-frontend-k8s-service:8080",
		CurlOptions: tools.CurlOpts{Timeout: 10},
	},
	ValidatorRetry: frame2.RetryOptions{
		Allow: 10,
	},
}

var prvBackWorks = frame2.Step{
	Doc: "Validate backend from prv",
	Validator: validate.Curl{
		Namespace:   prv,
		Url:         "http://hello-world-backend-k8s-service:8080/api/hello",
		CurlOptions: tools.CurlOpts{Timeout: 10},
	},
	ValidatorRetry: frame2.RetryOptions{
		Allow: 50,
	},
}

var prvFrontWorks = frame2.Step{
	Doc: "Validate frontend from prv",
	Validator: validate.Curl{
		Namespace:   prv,
		Url:         "http://hello-world-frontend-k8s-service:8080",
		CurlOptions: tools.CurlOpts{Timeout: 10},
	},
	ValidatorRetry: frame2.RetryOptions{
		Allow: 50,
	},
}

var tests = frame2.TestRun{
	Name:   "test-314",
	Runner: runner,
	Setup: []frame2.Step{
		{
			Doc: "Segment setup",
			Modify: walk.SegmentSetup{
				Namespace: pub,
			},
		}, {
			Doc: "create frontend service",
			Modify: execute.K8SServiceCreate{
				Namespace: pub,
				Name:      "hello-world-frontend-k8s-service",
				Selector:  map[string]string{"app": "hello-world-frontend"},
				Labels:    map[string]string{"app": "hello-world-frontend"},
				Ports:     []int32{8080},
			},
		}, {
			Doc: "create backend service",
			Modify: execute.K8SServiceCreate{
				Namespace: prv,
				Name:      "hello-world-backend-k8s-service",
				Selector:  map[string]string{"app": "hello-world-backend"},
				Labels:    map[string]string{"app": "hello-world-backend"},
				Ports:     []int32{8080},
			},
		},
	},
	Teardown: []frame2.Step{
		{
			Doc: "Teardown",
			Modify: walk.SegmentTeardown{
				Step: frame2.Step{Namespace: pub},
			},
		},
	},
	MainSteps: []frame2.Step{
		{
			Name: "pre-checks",
			Doc:  "Just ensuring the initial setup is good",
			Substeps: []*frame2.Step{
				&pubFrontWorks,
				&prvBackWorks,
			},
		}, {
			Name: "repeat",
			Doc:  "Repeat the same test many times",
			SubstepRetry: frame2.RetryOptions{
				Ensure: 30,
			},
			Substep: &frame2.Step{
				Name: "number",
				Substeps: []*frame2.Step{
					{
						Doc: "Annotating frontend for Skupper",
						Modify: execute.K8SServiceAnnotate{
							Namespace: pub,
							Name:      "hello-world-frontend-k8s-service",
							Annotations: map[string]string{
								types.ProxyQualifier:   "http",
								types.AddressQualifier: "hello-world-frontend-k8s-service",
							},
						},
					}, {
						Doc: "Annotating backend for Skupper",
						Modify: execute.K8SServiceAnnotate{
							Namespace: prv,
							Name:      "hello-world-backend-k8s-service",
							Annotations: map[string]string{
								types.ProxyQualifier:   "http",
								types.AddressQualifier: "hello-world-backend-k8s-service",
							},
						},
					},
					&prvFrontWorks,
					&pubBackWorks,
					&pubFrontWorks,
					&prvBackWorks,
					{
						Doc: "Removing annotation from frontend Skupper",
						Modify: execute.K8SServiceRemoveAnnotation{
							Namespace: pub,
							Name:      "hello-world-frontend-k8s-service",
							Annotations: []string{
								types.ProxyQualifier,
								types.AddressQualifier,
							},
						},
					}, {
						Doc: "Removing annotation from backend Skupper",
						Modify: execute.K8SServiceRemoveAnnotation{
							Namespace: prv,
							Name:      "hello-world-backend-k8s-service",
							Annotations: []string{
								types.ProxyQualifier,
								types.AddressQualifier,
							},
						},
					},
					&pubFrontWorks,
					&prvBackWorks,
					&pubBackDoesntWork,
					&prvFrontDoesntWork,
				},
			},
		},
	},
}
