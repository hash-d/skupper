package frame2

import (
	"fmt"
	"log"
	"testing"

	"github.com/skupperproject/skupper/test/utils/base"
)

type TestRun struct {
	Name      string
	Setup     []Stepper
	Teardown  []Stepper
	MainSteps []Stepper
	Runner    *base.ClusterTestRunnerBase
}

func processStep(step interface{}) error {
	log.Printf("hehehe %T", step)
	switch step := step.(type) {
	case Validator:
		log.Printf("hohoho")
		_, err := Retry{
			Fn:      step.Run,
			Options: step.GetRetryOptions(),
		}.Run()
		return err
	case Stepper:
		log.Printf("wha")
		return step.Run()
	default:
		return fmt.Errorf("Invalid step type: %T", step)
	}
}

func processValidate(step interface{}) error {
	log.Printf("Type: %T", step)
	v, ok := step.(Validator)
	if !ok {
		return fmt.Errorf("non-validate")
	}
	_, err := Retry{
		Fn:      v.Run,
		Options: v.GetRetryOptions(),
	}.Run()
	return err
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
		if err := processValidate(step); err != nil {
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
