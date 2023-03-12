//go:build reproducers
// +build reproducers

package main

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/topology/topologies"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func Test984(t *testing.T) {
	runner := frame2.Run{T: t}
	runnerBase := base.ClusterTestRunnerBase{}

	var topologySimple topology.Basic
	topologySimple = &topologies.Simplest{
		Name:           "test-984",
		TestRunnerBase: &runnerBase,
	}

	topologyPhase := frame2.Phase{
		Doc: "Create the o-o topology map",
		Setup: []frame2.Step{
			{
				Modify: topologySimple,
			},
		},
	}
	topologyPhase.Run()

	envPhase := frame2.Phase{
		Doc:    "Create the actual topology",
		Runner: &runner,
		Setup: []frame2.Step{
			{
				Modify: environment.HelloWorld{
					Runner:       &runner,
					Topology:     &topologySimple,
					AutoTearDown: true,
				},
			},
		},
	}
	envPhase.Run()

	pub, err := topologySimple.Get(topology.Public, 1)
	assert.Assert(t, err)

	testPhase := frame2.Phase{
		Runner: &runner,
		Name:   "add-second-port-pair",
		Doc:    "Expose a deployment via annotations, with a single port pair.  Then change the annotation to contain two pairs, and check for service-controller restart",
		Setup: []frame2.Step{{
			Doc: "Initial annotation, with a single port pair, then ensure the service was created",
			Modify: execute.K8SDeploymentAnnotate{
				Namespace: pub,
				Name:      "hello-world-frontend",
				Annotations: map[string]string{
					"skupper.io/address": "hello-world-frontend",
					"skupper.io/port":    "8080:8080",
					"skupper.io/proxy":   "http",
				},
			},
			Validator: validate.SkupperService{
				Namespace: pub,
				Name:      "hello-world-frontend",
			},
			ValidatorRetry: frame2.RetryOptions{
				Allow: 10,
			},
		}},
		MainSteps: []frame2.Step{{
			Name: "check-for-failure",
			Doc:  "Change the annotation to contain two pairs, and monitor the container",
			Modify: execute.K8SDeploymentAnnotate{
				Namespace: pub,
				Name:      "hello-world-frontend",
				Annotations: map[string]string{
					"skupper.io/port": "8080:8080,80:8080",
				},
			},
			Validator: validate.Container{
				Namespace:    pub,
				PodSelector:  validate.ServiceControllerSelector,
				RestartCount: 0,
				RestartCheck: true,
				StatusCheck:  true,
			},
			ValidatorRetry: frame2.RetryOptions{
				Ensure: 30,
			},
		}, {
			Name: "check-if-restored",
			Doc:  "After the previous step, regardless whether it failed; does the service controller eventually come back up?  And is the change active?",
			Validators: []frame2.Validator{
				validate.Container{
					Namespace:   pub,
					PodSelector: validate.ServiceControllerSelector,
					StatusCheck: true,
				},
				validate.Curl{
					Namespace: pub,
					Url:       "http://hello-world-frontend:80",
				},
			},
			ValidatorRetry: frame2.RetryOptions{
				Allow:  600,
				Ensure: 100,
			},
		}},
	}
	testPhase.Run()

}
