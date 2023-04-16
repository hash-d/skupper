//go:build meta_test
// +build meta_test

package template

import (
	"fmt"
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/deploy"
	"github.com/skupperproject/skupper/test/frame2/disruptors"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"gotest.tools/assert"
)

func TestPatientPortalTemplate(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}
	defer r.Finalize()

	r.AllowDisruptors(
		[]frame2.Disruptor{
			&disruptors.UpgradeAndFinalize{},
		},
	)

	env := environment.PatientPortalDefault{
		Runner:       r,
		AutoTearDown: true,
	}

	setup := frame2.Phase{
		Runner: r,
		Name:   "Patient Portal setup",
		Doc:    "Deploy Patient Portal on the default topology",
		Setup: []frame2.Step{
			{
				Doc:    "Deploy Patient Portal",
				Modify: &env,
			},
		},
	}

	assert.Assert(t, setup.Run())

	front_ns, err := env.TopoReturn.Get(topology.Public, 1)
	if err != nil {
		t.Fatalf(fmt.Sprintf("failed to get pub-1: %v", err))
	}

	main := frame2.Phase{
		Runner: r,
		Name:   "sample-tests",
		Doc:    "Replace these by your modifications and tests",
		MainSteps: []frame2.Step{
			{
				Validators: []frame2.Validator{
					&deploy.PatientValidatePayment{
						Runner:    r,
						Namespace: front_ns,
					},
				},
				ValidatorFinal: true,
			},
		},
	}

	assert.Assert(t, main.Run())

	// Teardown: for the template, all tear down is automatic.
	// If specific tear downs from the main steps are required,
	// create a new phase and specify them.

}
