package fixes

import (
	"testing"

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

var tests = frame2.TestRun{
	Name: "test-314",
	Setup: []frame2.Stepper{
		walk.SegmentSetup{
			Step: frame2.Step{Namespace: pub},
		},
		execute.K8SServiceCreate{
			Execute:  frame2.Execute{Step: frame2.Step{Namespace: pub}},
			Name:     "hello-world-frontend-k8s-service",
			Selector: map[string]string{"app": "hello-world-frontend"},
			Labels:   map[string]string{"app": "hello-world-frontend"},
			Ports:    []int32{8080},
		},
		execute.K8SServiceCreate{
			Execute:  frame2.Execute{Step: frame2.Step{Namespace: prv}},
			Name:     "hello-world-backend-k8s-service",
			Selector: map[string]string{"app": "hello-world-backend"},
			Labels:   map[string]string{"app": "hello-world-backend"},
			Ports:    []int32{8080},
		},
	},
	Teardown: []frame2.Stepper{
		walk.SegmentTeardown{
			Step: frame2.Step{Namespace: pub},
		},
	},
	MainSteps: []frame2.Stepper{
		validate.Curl{
			Validate: frame2.Validate{
				Step: frame2.Step{Namespace: pub},
				RetryOptions: frame2.RetryOptions{
					Allow: 10,
				},
			},
			Url:         "http://hello-world-frontend-k8s-service:8080",
			CurlOptions: tools.CurlOpts{Timeout: 10},
		},
		validate.Curl{
			Validate:    frame2.Validate{Step: frame2.Step{Namespace: prv}},
			Url:         "http://hello-world-backend-k8s-service:8080",
			CurlOptions: tools.CurlOpts{Timeout: 10},
		},
	},
	Runner: runner,
}
