package frame2

import (
	"fmt"
	"log"
	"testing"

	"github.com/skupperproject/skupper/test/utils/base"
)

type TestRun struct {
	Name      string
	Doc       string
	Setup     []Step
	Teardown  []Step
	MainSteps []Step
	Runner    *base.ClusterTestRunnerBase
	teardowns []Executor
}

func processStep(t *testing.T, step Step) error {
	// TODO: replace [TR] with own logger with Prefix?
	var err error
	if step.Name != "" {
		_ = t.Run(step.Name, func(t *testing.T) {
			log.Printf("[TR] Doc: %v", step.Doc)
			processErr := processStep_(t, step)
			if processErr != nil {
				// This make it easier to find the failures in log files
				log.Printf("[TR] test %q failed", t.Name())
				// For named tests, we do not return the error up; we
				// just mark it as a failed test
				t.Errorf("test failed: %v", processErr)
			}
		})
		//		if !ret {
		//			err = fmt.Errorf("test failed: %v", err)
		//		}
	} else {
		log.Printf("[TR] Running step TBD# doc %q", step.Doc)
		err = processStep_(t, step)
	}
	return err

}
func processStep_(t *testing.T, step Step) error {
	if step.Modify != nil {
		log.Printf("[TR] Modifier %T", step.Modify)
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

func (tr *TestRun) Run(t *testing.T) error {

	if tr.Name == "" {
		t.Fatal("test name must be defined")
	}

	if tr.Runner == nil {
		tr.Runner = &base.ClusterTestRunnerBase{}
	}

	// TODO: allow for optional interface.  If the step also implements Teardown(),
	// execute it and add its result to the teardown list.
	log.Printf("Starting setup")
	for _, step := range tr.Setup {
		//		if step.Modify != nil {
		//			if step, ok := step.Modify.(TearDowner); ok {
		//				tr.teardowns = append(tr.teardowns, step.TearDown())
		//			}
		//		}
		if err := processStep(t, step); err != nil {
			t.Errorf("setup failed: %v", err)
			return err
		}
	}
	log.Printf("Starting main steps")
	for _, step := range tr.MainSteps {
		if err := processStep(t, step); err != nil {
			t.Errorf("test failed: %v", err)
			tr.Runner.DumpTestInfo(tr.Name)
			break
		}
	}

	log.Printf("Starting auto-teardown")
	for _, td := range tr.teardowns {
		if err := td.Execute(); err != nil {
			t.Errorf("auto-teardown failed: %v", err)
			// We do not return here; we keep going doing whatever
			// teardown we can
		}
	}

	log.Printf("Starting teardown")
	for _, step := range tr.Teardown {
		if err := processStep(t, step); err != nil {
			t.Errorf("teardown failed: %v", err)
			// We do not return here; we keep going doing whatever
			// teardown we can
		}
	}
	return nil
}
