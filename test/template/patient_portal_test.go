//go:build meta_test
// +build meta_test

package template

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"gotest.tools/assert"
)

func TestPatientPortalTemplate(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}

	setup := frame2.Phase{
		Runner: r,
		Name:   "Patient Portal setup",
		Doc:    "Deploy Patient Portal on the default topology",
		Setup: []frame2.Step{
			{
				Name: "Deploy Patient Portal",
				Doc:  "Deploy Patient Portal",
				Modify: environment.PatientPortalDefault{
					Runner:       r,
					AutoTearDown: true,
				},
			},
		},
	}

	assert.Assert(t, setup.Run())

	main := frame2.Phase{
		Runner: r,
		Name:   "Replace me",
		Doc:    "Here goes the steps of the actual test",
		MainSteps: []frame2.Step{
			{
				Modify: execute.Function{
					Fn: func() error {
						time.Sleep(time.Minute * 10)
						return nil
					},
				},
			},
		},
	}

	assert.Assert(t, main.Run())

	// Teardown: for the template, all tear down is automatic.
	// If specific tear downs from the main steps are required,
	// create a new phase and specify them.

}
