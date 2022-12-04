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

func processStep(step Step) error {
	log.Printf("Running step doc %q", step.Doc)
	if step.Modify != nil {
		err := step.Modify.Execute()
		if err != nil {
			return fmt.Errorf("validate step failed: %w", err)
		}
	}
	if step.Validator != nil {
		_, err := Retry{
			Fn:      step.Validator.Validate,
			Options: step.ValidatorRetry,
		}.Run()
		return err
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

	log.Printf("Starting setup")
	for _, step := range tr.Setup {
		if err := processStep(step); err != nil {
			return err
		}
	}
	log.Printf("Starting main steps")
	for _, step := range tr.MainSteps {
		if err := processStep(step); err != nil {
			t.Errorf("test failed: %v", err)
			tr.Runner.DumpTestInfo(tr.Name)
			break
		}
	}
	log.Printf("Starting teardown")
	for _, step := range tr.Teardown {
		if err := processStep(step); err != nil {
			return err
		}
	}
	return nil
}
