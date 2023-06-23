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
	defer r.Report()
	defer r.Finalize()

	r.AllowDisruptors(
		[]frame2.Disruptor{
			&disruptors.UpgradeAndFinalize{},
			&disruptors.MixedVersionVan{},
			&disruptors.DeploymentConfigBlindly{},
			&disruptors.NoConsole{},
			&disruptors.NoFlowCollector{},
			&disruptors.NoHttp{},
			&disruptors.ConsoleOnAll{},
			&disruptors.FlowCollectorOnAll{},
			&disruptors.MinAllows{},
		},
	)

	env := environment.PatientPortalDefault{
		AutoTearDown:  true,
		EnableConsole: true,
	}

	setup := frame2.Phase{
		Runner: r,
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

	// In case you're running with disruptors, the monitors not only inform when connectivity
	// was disrupted, but also help with applications re-establishing their connections (on
	// a pool, for example) after a disruption, and before the actual test.
	monitorPhase := frame2.Phase{
		Runner: r,
		Setup: []frame2.Step{
			{
				Doc: "Install Patient Portal monitors",
				Modify: &frame2.DefaultMonitor{
					Validators: map[string]frame2.Validator{
						"frontend-health": &deploy.PatientFrontendHealth{
							Namespace: front_ns,
						},
						"database-ping": &deploy.PatientDbPing{
							Namespace: front_ns,
						},
						"payment-token": &deploy.PatientValidatePayment{
							Namespace: front_ns,
						},
					},
				},
			},
		},
	}
	assert.Assert(t, monitorPhase.Run())

	main := frame2.Phase{
		Runner: r,
		Name:   "sample-tests",
		Doc:    "Replace these by your modifications and tests",
		MainSteps: []frame2.Step{
			{
				Validators: []frame2.Validator{
					&deploy.PatientValidatePayment{
						Namespace: front_ns,
					},
					&deploy.PatientFrontendHealth{
						Namespace: front_ns,
					},
					&deploy.PatientDbPing{
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
