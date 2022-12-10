//go:build meta_test
// +build meta_test

package frame2_test

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func TestPlayground(t *testing.T) {

	var runner = &base.ClusterTestRunnerBase{}

	var tests = frame2.Phase{
		Name: "test-playground",
		Doc:  "play with it",
		Setup: []frame2.Step{
			{
				Doc:    "Please succeed",
				Modify: execute.Success{},
			},
		},
		Teardown: []frame2.Step{},
		MainSteps: []frame2.Step{
			{
				Name: "dummy",
				Doc:  "Dummy testing",
				Validator: &validate.Dummy{
					Results: []error{io.EOF, nil, nil, io.EOF, nil, io.EOF, nil},
				},
				ValidatorRetry: frame2.RetryOptions{
					Ignore:   2,
					Retries:  1,
					Interval: time.Microsecond,
				},
			},
			{
				Name: "sub",
				Doc:  "Testing substeps",
				Substep: &frame2.Step{
					Validator: &validate.Dummy{
						Results: []error{io.EOF, nil, io.EOF, nil, nil},
					},
				},
				SubstepRetry: frame2.RetryOptions{
					Allow:    1,
					Ignore:   2,
					Retries:  1,
					Ensure:   2,
					Interval: time.Microsecond,
				},
			},
		},
		BaseRunner: runner,
	}
	assert.Assert(t, tests.RunT(t))
}

func TestEmpty(t *testing.T) {

	//runner := frame2.Run{T: t}

	tests := frame2.Phase{
		Name: "Test Empty",
	}
	tests.RunT(t)

}

func TestSimplest(t *testing.T) {
	tests := frame2.Phase{
		Name: "Simplest",
		MainSteps: []frame2.Step{
			{
				Modify: execute.Success{},
			},
		},
	}
	tests.RunT(t)
}

func TestTwoPhases(t *testing.T) {

	runner := frame2.Run{T: t}

	phase1 := frame2.Phase{
		Runner: &runner,
		Name:   "Phase1",
		MainSteps: []frame2.Step{
			{
				Doc:    "Phase1",
				Modify: execute.Success{},
			},
		},
	}
	phase1.Run()

	phase2 := frame2.Phase{
		Runner: &runner,
		Name:   "Phase2",
		MainSteps: []frame2.Step{
			{
				Doc:    "Phase2",
				Modify: execute.Success{},
			},
		},
	}
	phase2.Run()

	for i := 1; i < 3; i++ {

		phase3 := frame2.Phase{
			Runner: &runner,
			Name:   "Repeating phase",
			MainSteps: []frame2.Step{
				{
					Doc:    fmt.Sprintf("Phase3.%d", i),
					Modify: execute.Success{},
				},
			},
		}
		phase3.Run()
	}

	innerPhase := frame2.Phase{
		Runner: &runner,
		Name:   "Inner phase",
		MainSteps: []frame2.Step{
			{
				Doc:    "InnerPhase",
				Modify: execute.Success{},
			},
		},
	}

	phase4 := frame2.Phase{
		Runner: &runner,
		Name:   "Phase4",
		MainSteps: []frame2.Step{
			{
				Doc:    "Phase 4",
				Modify: innerPhase,
			},
		},
	}
	phase4.Run()

	var checked bool

	phase5 := frame2.Phase{
		Runner: &runner,
		Name:   "Closure",
		MainSteps: []frame2.Step{
			{
				Doc:  "Closure 1: set",
				Name: "Compo",
				Modify: execute.Function{
					Fn: func() error {
						if checked {
							return fmt.Errorf("Checked started with true!")
						}
						checked = true
						return nil
					},
				},
			}, {
				Doc:  "Closure 2: get",
				Name: "Compo",
				Modify: execute.Function{
					Fn: func() error {
						if !checked {
							return fmt.Errorf("Checked was not changed!")
						}
						return nil
					},
				},
			},
		},
	}
	phase5.Run()

	original := "World!"

	phase6 := frame2.Phase{
		Runner: &runner,
		Name:   "Composition",
		MainSteps: []frame2.Step{
			{
				Doc: "Calling composition",
				Modify: Composed{
					Runner:    &runner,
					Argument:  "Hello",
					Reference: &original,
				},
			},
		},
	}
	phase6.Run()

	// closure, composed

}

type Composed struct {
	Runner    *frame2.Run
	Argument  string
	Reference *string
}

func (c Composed) Execute() error {
	compoPhase1 := frame2.Phase{
		Runner: c.Runner,
		Name:   "CompoPhase1",
		MainSteps: []frame2.Step{
			{
				Doc: "Print start",
				Modify: execute.Print{
					Message: "Got values %q and %q",
					Data:    []interface{}{c.Argument, *c.Reference},
				},
			}, {
				Doc: "Modify",
				Modify: execute.Function{
					Fn: func() error {
						newValue := "Changed!"
						c.Reference = &newValue
						return nil
					},
				},
			},
		},
	}
	compoPhase1.Run()

	compoPhase2 := frame2.Phase{
		Runner: c.Runner,
		Name:   "CompoPhase2",
		MainSteps: []frame2.Step{
			{
				Doc: "Print final",
				Modify: execute.Print{
					Message: "Got values %q and %q",
					Data:    []interface{}{c.Argument, *c.Reference},
				},
			},
		},
	}
	compoPhase2.Run()
	return nil
}
