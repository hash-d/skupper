package disruptors

import (
	"os"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

// At the end of the test, before the tear down, upgrade all
// sites and then re-run all tests marked as final
//
// This is a very basic upgrade test; it's cheap and simple
type UpgradeAndFinalize struct {
	targets []*base.ClusterContext
	useNew  bool
}

func (u UpgradeAndFinalize) DisruptorEnvValue() string {
	return "UPGRADE_AND_FINALIZE"
}

func (u *UpgradeAndFinalize) PreFinalizerHook(runner *frame2.Run) error {
	var steps []frame2.Step
	u.useNew = true

	for _, t := range u.targets {
		steps = append(steps, frame2.Step{
			Modify: execute.SkupperUpgrade{
				Runner:    runner,
				Namespace: t,
				Wait:      time.Minute * 10,
			},
		})
	}
	phase := frame2.Phase{
		Runner:    runner,
		MainSteps: steps,
	}
	return phase.Run()
}

func (u *UpgradeAndFinalize) Inspect(step frame2.Step, phase frame2.Phase) {
	if step, ok := step.Modify.(execute.SkupperUpgradable); ok {
		u.targets = append(u.targets, step.SkupperUpgradable())
	}
	if step, ok := step.Modify.(execute.SkupperCliPathSetter); ok {
		if !u.useNew {
			path := os.Getenv("SKUPPER_TEST_OLD_BIN")
			if path == "" {
				panic("Disruptor UPGRADE_AND_FINALIZE requested, but no SKUPPER_TEST_OLD_BIN config")
			}
			step.SetSkupperCliPath(path)
		}
	}
}
