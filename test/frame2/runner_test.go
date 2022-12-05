//go:build meta_test
// +build meta_test

package frame2_test

import (
	"io"
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func TestTestRunner(t *testing.T) {
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
			Name: "dummy",
			Doc:  "Dummy testing",
			Validator: &validate.Dummy{
				Results: []error{io.EOF, nil, nil, io.EOF, nil, io.EOF, nil},
			},
			ValidatorRetry: frame2.RetryOptions{
				Ignore:   2,
				Retries:  1,
				Interval: time.Microsecond,
			},
		},
		{
			Name: "sub",
			Doc:  "Testing substeps",
			Substep: &frame2.Step{
				Validator: &validate.Dummy{
					Results: []error{io.EOF, nil, nil, io.EOF, nil, io.EOF, nil},
				},
			},
			SubstepRetry: frame2.RetryOptions{
				Ignore:   2,
				Retries:  1,
				Ensure:   2,
				Interval: time.Microsecond,
			},
		},
	},
	Runner: runner,
}
