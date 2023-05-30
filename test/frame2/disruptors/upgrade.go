package disruptors

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Sort the targets according to some strategy, configured on
// SKUPPER_TEST_UPGRADE_STRATEGY.  If none set, return the
// target list unchanged
func sortTargets(targets []*base.ClusterContext) []*base.ClusterContext {

	var ret []*base.ClusterContext
	var invert bool
	var strategy frame2.TestUpgradeStrategy

	envValue := os.Getenv(frame2.ENV_UPGRADE_STRATEGY)

	s := strings.SplitN(envValue, ":", 2)
	strategy = frame2.TestUpgradeStrategy(s[0])
	if strategy == "" {
		strategy = frame2.UPGRADE_STRATEGY_CREATION
	}
	if len(s) > 1 {
		if s[1] == string(frame2.UPGRADE_STRATEGY_INVERSE) {
			invert = true
		} else {
			panic(fmt.Sprintf("invalid option to SKUPPER_TEST_UPGRADE_STRATEGY: %v", s[1]))
		}
	}

	switch strategy {
	case frame2.UPGRADE_STRATEGY_CREATION:
		ret = targets[:]
	default:
		panic(fmt.Sprintf("invalid upgrade strategy: %v", strategy))
	}

	if invert {
		lenRet := len(ret)
		for i := 0; i < lenRet/2; i++ {
			ret[i], ret[lenRet-i-1] = ret[lenRet-i-1], ret[i]
		}
	}

	return ret
}

// At the end of the test, before the tear down, upgrade all
// sites and then re-run all tests marked as final
//
// This is a very basic upgrade test; it's cheap and simple
//
// The upgrade strategy can be defined on the environment
// variable SKUPPER_TEST_UPGRADE_STRATEGY.
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

	targets := sortTargets(u.targets)

	for _, t := range targets {
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

func (u *UpgradeAndFinalize) Inspect(step *frame2.Step, phase *frame2.Phase) {
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
