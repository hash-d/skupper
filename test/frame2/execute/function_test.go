package execute

import (
	"errors"
	"log"
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
)

func TestFunction(t *testing.T) {
	tests := frame2.Phase{
		Name: "TestFunction",
		MainSteps: []frame2.Step{
			{
				Name: "func-ok",
				Modify: Function{
					Fn: func() error {
						log.Printf("Hello")
						return nil
					},
				},
			}, {
				Name: "func-fail",
				Modify: Function{
					Fn: func() error {
						return errors.New("failed!")
					},
				},
				// This does not help, as it allows only for
				// validation errors; modify steps are not expected
				// to fail
				ExpectError: true,
			},
		},
	}
	tests.RunT(t)
}
