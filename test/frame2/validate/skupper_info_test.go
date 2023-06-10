package validate_test

import (
	"fmt"
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"gotest.tools/assert"
)

func TestSkupperInfo(t *testing.T) {

	run := &frame2.Run{
		T: t,
	}

	installCurrent := environment.JustSkupperDefault{
		Name:    "frame2-skupper-info-test",
		Console: true,
		//AutoTearDown: true,
	}
	//installOld := environment.JustSkupperDefault{ }

	SetupPhase := frame2.Phase{
		Runner: run,
		Doc:    "Setup a namespace with Skupper, no other deployments",
		Setup: []frame2.Step{
			{
				Modify: &installCurrent,
			},
		},
	}
	assert.Assert(t, SetupPhase.Run())

	namespace, err := installCurrent.Topo.Get(topology.Public, 1)
	assert.Assert(t, err)

	getInfoCurrent := validate.SkupperInfo{
		Namespace: namespace,
	}

	infoPhase := frame2.Phase{
		Runner: run,
		Doc:    "Get Skupper information, to compare to manifest.json",
		MainSteps: []frame2.Step{
			{
				Validator: &getInfoCurrent,
			},
		},
	}
	assert.Assert(t, infoPhase.Run())

	testPhase := frame2.Phase{
		Runner: run,
		Doc:    "Compare manifest.json to Skupper info acquired priorly",
		MainSteps: []frame2.Step{
			{
				Modify: execute.Function{
					Fn: func() error {
						if getInfoCurrent.Result.HasRouter {
							return nil
						}
						return fmt.Errorf("The namespace %q has no skupper-router, so we can't consider it for manifest check", namespace.Namespace)
					},
				},
				Validator: &validate.SkupperManifest{
					Expected: getInfoCurrent.Result.Images,
				},
			},
		},
	}
	assert.Assert(t, testPhase.Run())

}
