//go:build meta_test
// +build meta_test

package frame2_test

import (
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"gotest.tools/assert"
)

func TestPlayground(t *testing.T) {

	//var runner = &base.ClusterTestRunnerBase{}

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
		//BaseRunner: runner,
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

type AutoDestruct struct {
}

func (a AutoDestruct) Execute() error {
	log.Println("Autodestruct active")
	return nil
}

func (a AutoDestruct) TearDown() frame2.Executor {
	return execute.Print{
		Message: "Destroyed!",
	}
}

func TestAutoTearDown(t *testing.T) {
	runner := frame2.Run{T: t}

	test := frame2.Phase{
		Name:   "Phase1",
		Runner: &runner,
		Setup: []frame2.Step{
			{
				Modify: AutoDestruct{},
			},
		},
	}
	test.Run()

}

// This is used for TestInner
type SimpleComposed struct {
	Runner *frame2.Run
}

func (s SimpleComposed) Execute() error {
	phase := frame2.Phase{
		Runner: s.Runner,
		Name:   "Composed-Inner",
		Doc:    "The phase within the composed Executor",
		MainSteps: []frame2.Step{
			{
				Doc:    "The step within the composed Executor",
				Modify: execute.Success{},
			},
		},
	}
	return phase.Run()
}

func TestInner(t *testing.T) {

	runner := frame2.Run{
		T:   t,
		Doc: "Tests different types of child executions",
	}

	testSubsteps := frame2.Phase{
		Name:   "Substeps",
		Doc:    "execute.Success is executed within a substep on a top-level Phase",
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Doc: "The containing step",
				Substeps: []*frame2.Step{
					{
						Doc:    "Unnamed substep 1",
						Modify: execute.Success{},
					}, {
						Name:   "Substep-Inner-1",
						Doc:    "The inner substep",
						Modify: execute.Success{},
					}, {
						Doc:    "Unnamed substep 2",
						Modify: execute.Success{},
					}, {
						Name:   "Substep-Inner-2",
						Doc:    "The inner substep 2",
						Modify: execute.Success{},
					},
				},
			},
		},
	}
	testSubsteps.Run()

	testInnerPhase := frame2.Phase{
		Name:   "Inner-phases",
		Doc:    "execute.Success is executed within an inner phase (a phase used as a Modify in a top level Phase)",
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Doc: "The containing step for the unnamed inner phase 1",
				Modify: frame2.Phase{
					Doc: "The unnamed inner phase 1",
					MainSteps: []frame2.Step{
						{
							Doc:    "The inner phase's step",
							Modify: execute.Success{},
						},
					},
				},
			}, {
				Doc: "The containing step for the named inner phase 1",
				Modify: frame2.Phase{
					//			Runner: &runner,
					Name: "Phase-Inner-1",
					Doc:  "The inner phase 1",
					MainSteps: []frame2.Step{
						{
							Doc:    "The inner phase's step",
							Modify: execute.Success{},
						},
					},
				},
			}, {
				Doc: "The containing step for the unnamed inner phase 2",
				Modify: frame2.Phase{
					Doc: "The unnamed inner phase 2",
					MainSteps: []frame2.Step{
						{
							Doc:    "The inner phase's step",
							Modify: execute.Success{},
						},
					},
				},
			}, {
				Doc: "The containing step for the named inner phase 2",
				Modify: frame2.Phase{
					//			Runner: &runner,
					Name: "Phase-Inner-2",
					Doc:  "The inner phase 2",
					MainSteps: []frame2.Step{
						{
							Doc:    "The inner phase's step",
							Modify: execute.Success{},
						},
					},
				},
			},
		},
	}
	testInnerPhase.Run()

	testComposed := frame2.Phase{
		Name:   "Inner-Composed",
		Doc:    "execute.Success is executed in a composed Executor (an Executor that has its own internal phase, connected to the same Runner as the parent Phase)",
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Doc: "The containing step",
				Modify: SimpleComposed{
					Runner: &runner,
				},
			},
		},
	}
	testComposed.Run()

}
