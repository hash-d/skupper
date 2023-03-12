//go:build fixes
// +build fixes

package fixes

import (
	"testing"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/tools"
	"gotest.tools/assert"
)

func TestSkupper314(t *testing.T) {

	runner := &base.ClusterTestRunnerBase{}
	f2runner := frame2.Run{
		T: t,
	}

	var topo topology.Basic
	topo = &topologies.Simplest{
		Name:           "skupper-314",
		TestRunnerBase: runner,
	}

	topoPhase := frame2.Phase{
		Runner: &f2runner,
		Setup: []frame2.Step{
			{
				Modify: environment.HelloWorld{
					Runner:       &f2runner,
					Topology:     &topo,
					AutoTearDown: true,
				},
			},
		},
	}
	topoPhase.Run()

	pub, err := topo.Get(topology.Public, 1)
	assert.Assert(t, err)
	prv, err := topo.Get(topology.Private, 1)
	assert.Assert(t, err)

	var pubBackDoesntWork = frame2.Step{
		Doc: "Validate backend from pub doesnt work",
		Validator: validate.Curl{
			Namespace:   pub,
			Url:         "http://hello-world-backend-k8s-service:8080/api/hello",
			CurlOptions: tools.CurlOpts{Timeout: 10},
		},
		ValidatorRetry: frame2.RetryOptions{
			Allow:  10,
			Ignore: 10,
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
			Allow:  10,
			Ignore: 10,
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

	var tests = frame2.Phase{
		Name:   "test-314",
		Runner: &f2runner,
		Setup: []frame2.Step{
			{
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
									types.ProxyQualifier:   "tcp",
									types.AddressQualifier: "hello-world-frontend-k8s-service",
								},
							},
						}, {
							Doc: "Annotating backend for Skupper",
							Modify: execute.K8SServiceAnnotate{
								Namespace: prv,
								Name:      "hello-world-backend-k8s-service",
								Annotations: map[string]string{
									types.ProxyQualifier:   "tcp",
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
						// First try what's supposed to change...
						&pubBackDoesntWork,
						&prvFrontDoesntWork,
						// Then what's supposed to remain working
						&pubFrontWorks,
						&prvBackWorks,
					},
				},
			},
		},
	}
	assert.Assert(t, tests.Run())
}
