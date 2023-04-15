package main

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"gotest.tools/assert"
)

func TestPatientPortal(t *testing.T) {

	r := frame2.Run{
		T: t,
	}

	setup := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Doc: "Stand up a Patient Portal deployment, on the platform topology",
			},
		},
	}

	assert.Assert(t, setup.Run())
}
