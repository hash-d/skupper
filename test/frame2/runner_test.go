//go:build meta_test
// +build meta_test

package frame2_test

import (
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func TestRunner(t *testing.T) {
	assert.Assert(t, tests.Run(t))
}

var runner = &base.ClusterTestRunnerBase{}

var pub = runner.GetPublicContextPromise(1)
var prv = runner.GetPrivateContextPromise(1)

var tests = frame2.TestRun{
	Name:     "test-314",
	Setup:    []frame2.Step{},
	Teardown: []frame2.Step{},
	MainSteps: []frame2.Step{
		{
			Validator: validate.Dummy{
				Results: []error{nil, nil, nil},
			},
		},
	},
	Runner: runner,
}
