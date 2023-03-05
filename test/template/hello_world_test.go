//go:build meta_test
// +build meta_test

package template

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/environment"
	"gotest.tools/assert"
)

func TestHelloWorldTemplate(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}

	setup := frame2.Phase{
		Runner: r,
		Name:   "Hello World setup",
		Doc:    "Deploy Hello World on the default topology",
		Setup: []frame2.Step{
			{
				Name: "Deploy Hello World",
				Doc:  "Deploy Hello World",
				// As an alternative, check HelloWorldN for HelloWorld on
				// an N topology, or just HelloWorld for a fully configurable
				// environment.
				Modify: environment.HelloWorldDefault{},
			},
		},
	}

	assert.Assert(t, setup.Run())

	main := frame2.Phase{
		Runner: r,
		Name:   "Replace me",
		Doc:    "Here goes the steps of the actual test",
	}

	assert.Assert(t, main.Run())

	// Teardown: for the template, all tear down is automatic.
	// If specific tear downs from the main steps are required,
	// create a new phase and specify them.

}
