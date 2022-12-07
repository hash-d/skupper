package frame2

import (
	"fmt"
	"log"
	"testing"

	"github.com/skupperproject/skupper/test/utils/base"
)

type TestRun struct {
	Name      string
	Setup     []Step
	Teardown  []Step
	MainSteps []Step
	Runner    *base.ClusterTestRunnerBase
}

func processStep(t *testing.T, step Step) error {
	var err error
	if step.Name != "" {
		_ = t.Run(step.Name, func(t *testing.T) {
			log.Printf("Doc: %v", step.Doc)
			processErr := processStep_(t, step)
			if processErr != nil {
				// This make it easier to find the failures in log files
				log.Printf("test %q failed", step.Name)
				// For named tests, we do not return the error up; we
				// just mark it as a failed test
				t.Errorf("test failed: %v", processErr)
			}
		})
		//		if !ret {
		//			err = fmt.Errorf("test failed: %v", err)
		//		}
	} else {
		log.Printf("Running step doc %q", step.Doc)
		err = processStep_(t, step)
	}
	return err

}
func processStep_(t *testing.T, step Step) error {
	if step.Modify != nil {
		err := step.Modify.Execute()
		if err != nil {
			return fmt.Errorf("modify step failed: %w", err)
		}
	}
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
		err := v.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (tr *TestRun) Run(t *testing.T) error {

	if tr.Name == "" {
		return fmt.Errorf("test name must be defined")
	}

	if tr.Runner == nil {
		tr.Runner = &base.ClusterTestRunnerBase{}
	}

	// TODO: allow for optional interface.  If the step also implements Teardown(),
	// execute it and add its result to the teardown list.
	log.Printf("Starting setup")
	for _, step := range tr.Setup {
		if err := processStep(t, step); err != nil {
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
	log.Printf("Starting teardown")
	for _, step := range tr.Teardown {
		if err := processStep(t, step); err != nil {
			return err
		}
	}
	return nil
}
