package frame2

import (
	"log"
	"testing"

	"github.com/skupperproject/skupper/test/utils/base"
)

type TestRun struct {
	Setup     []Stepper
	Teardown  []Stepper
	MainSteps []Stepper
	Runner    *base.ClusterTestRunnerBase
}

func (tr *TestRun) Run(t *testing.T) error {

	if tr.Runner == nil {
		tr.Runner = &base.ClusterTestRunnerBase{}
	}

	log.Printf("Starting setup")
	for _, step := range tr.Setup {
		if err := step.Run(); err != nil {
			return err
		}
	}
	log.Printf("Starting main steps")
	for _, step := range tr.MainSteps {
		if err := step.Run(); err != nil {
			break
		}
	}
	log.Printf("Starting teardown")
	for _, step := range tr.Teardown {
		if err := step.Run(); err != nil {
			return err
		}
	}
	return nil
}
