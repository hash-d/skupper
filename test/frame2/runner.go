package frame2

import (
	"fmt"
	"log"
	"testing"

	"github.com/skupperproject/skupper/test/utils/base"
)

type Run struct {
	T      *testing.T
	savedT *testing.T
}

type Phase struct {
	Name       string
	Doc        string
	Setup      []Step
	Teardown   []Step
	MainSteps  []Step
	BaseRunner *base.ClusterTestRunnerBase
	teardowns  []Executor
	Runner     *Run

	savedRunner *Run
	previousRun bool
}

func processStep(t *testing.T, step Step) error {
	// TODO: replace [R] with own logger with Prefix?
	var err error
	if step.Name != "" {
		// For a named test, run or fail, we work the same.  It's up to t to
		// mark it as failed
		_ = t.Run(step.Name, func(t *testing.T) {
			log.Printf("[R] Doc: %v", step.Doc)
			processErr := processStep_(t, step)
			if processErr != nil {
				// This makes it easier to find the failures in log files
				log.Printf("[R] test %q failed", t.Name())
				// For named tests, we do not return the error up; we
				// just mark it as a failed test
				t.Errorf("test failed: %v", processErr)
			}
		})
	} else {
		// TODO.  Add the step number (like 2.1.3)
		log.Printf("[R] Step doc %q", step.Doc)
		err = processStep_(t, step)
	}
	return err

}
func processStep_(t *testing.T, step Step) error {
	if step.Modify != nil {
		log.Printf("[R] Modifier %T", step.Modify)
		err := step.Modify.Execute()
		if err != nil {
			return fmt.Errorf("modify step failed: %w", err)
		}
	}

	// TODO here and elsewhere: join Substep and Substeps in a single
	// list and use just one code.
	if step.Substep != nil {
		_, err := Retry{
			Fn: func() error {
				return processStep(t, *step.Substep)
			},
			Options: step.SubstepRetry,
		}.Run()
		if err != nil {
			return fmt.Errorf("substep failed: %w", err)
		}
	}
	for _, subStep := range step.Substeps {
		_, err := Retry{
			Fn: func() error {
				return processStep(t, *subStep)
			},
			Options: subStep.SubstepRetry,
		}.Run()
		if err != nil {
			return fmt.Errorf("substep failed: %w", err)
		}

	}
	if step.Validator != nil {
		log.Printf("[TR] Validator %T", step.Validator)
		_, err := Retry{
			Fn:      step.Validator.Validate,
			Options: step.ValidatorRetry,
		}.Run()
		if step.ExpectError {
			if err == nil {
				return fmt.Errorf("Error expected but not realised")
			} else {
				return nil
			}
		}
		if err != nil {
			return err
		}
	}
	for _, v := range step.Validators {
		log.Printf("[TR] Validator %T", v)
		err := v.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// For a Phase that did not define a Run, this will create a Run
// and set its T accordingly
//
// This is only for the simplest case, when a single phase is
// required.
//
// If the Phase already had a Runner set, it will fail.
func (p *Phase) RunT(t *testing.T) error {
	if p.Runner == nil {
		p.Runner = &Run{
			T: t,
		}
		p.savedRunner = p.Runner
	} else {
		return fmt.Errorf("Phase.RunT configuration error: cannot reset the Runner")
	}
	return p.Run()
}

func (p Phase) Execute() error {
	return p.Run()
}

func (p *Phase) Run() error {
	err := p.run()
	if err != nil {
		if p.Runner.T != nil {
			p.Runner.T.Fatalf("Phase error: %v", err)
		} else {
			log.Printf("Phase error: %v", err)
		}
	}
	return err
}

func (p *Phase) run() error {

	// If the phase has no runner, let's create one, without a *testing.T.  This
	// allows the runner to be used disconneced from the testing module.  This
	// way, Actions can be composed using a Phase
	runner := p.Runner
	if runner == nil {
		p.Runner = &Run{}
		p.savedRunner = p.Runner
	}

	// The Runner is public; we do not want people messing with it in the middle
	// of a Run
	if p.previousRun && p.Runner != p.savedRunner {
		log.Printf("saved: %v, new: %v", p.savedRunner, p.Runner)
		return fmt.Errorf("Phase.Run configuration error: the Runner was changed")

	} else {
		p.savedRunner = p.Runner
	}
	t := runner.T
	//  The testing.T on the Runner is public.  We don't want people messing with
	//  it either.
	if p.previousRun {
		if t != runner.savedT {
			return fmt.Errorf("Phase.Run configuration error: the *testing.T inside the Runner was changed")
		}
	} else {
		p.savedRunner.savedT = t
		p.previousRun = true
	}

	if t != nil && p.Name == "" {
		t.Fatal("test name must be defined")
	}

	if t != nil && p.BaseRunner == nil {
		p.BaseRunner = &base.ClusterTestRunnerBase{}
	}

	// TODO: allow for optional interface.  If the step also implements Teardown(),
	// execute it and add its result to the teardown list.
	if len(p.Setup) > 0 {
		log.Printf("Starting setup")
		for _, step := range p.Setup {
			//		if step.Modify != nil {
			//			if step, ok := step.Modify.(TearDowner); ok {
			//				tr.teardowns = append(tr.teardowns, step.TearDown())
			//			}
			//		}
			if err := processStep(t, step); err != nil {
				if t != nil {
					t.Errorf("setup failed: %v", err)
				}
				return err
			}
		}
	}
	var savedError error
	if len(p.MainSteps) > 0 {
		log.Printf("Starting main steps")
		for _, step := range p.MainSteps {
			if err := processStep(t, step); err != nil {
				if t != nil {
					t.Errorf("test failed: %v", err)
					savedError = err
				}
				// TODO this should be pluggable
				//p.BaseRunner.DumpTestInfo(p.Name)
				break
			}
		}
	}

	if len(p.teardowns) > 0 {
		// TODO move this to t.Cleanup and make it depend on t != nil?
		log.Printf("Starting auto-teardown")
		for _, td := range p.teardowns {
			if err := td.Execute(); err != nil {
				if t != nil {
					t.Errorf("auto-teardown failed: %v", err)
				}
				// We do not return here; we keep going doing whatever
				// teardown we can
			}
		}
	}

	if len(p.Teardown) > 0 {
		log.Printf("Starting teardown")
		for i, step := range p.Teardown {
			if err := processStep(t, step); err != nil {
				if t == nil {
					log.Printf("Tear down step %d failed: %v", i, err)
					t.Errorf("teardown failed: %v", err)
				}
				// We do not return here; we keep going doing whatever
				// teardown we can
			}
		}
	}
	return savedError
}
