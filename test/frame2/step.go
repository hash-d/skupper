package frame2

import (
	"log"
	"os"
	"strings"

	"github.com/skupperproject/skupper/test/utils/base"
)

const EnvFrame2Verbose = "SKUPPER_TEST_FRAME2_VERBOSE"

type Step struct {
	Doc       string
	Namespace *base.ClusterContextPromise
	Name      string
	Level     int
	// Whether the step should always print logs
	// Even if false, logs will be done if SKUPPER_TEST_FRAME2_VERBOSE
	Verbose        bool
	PreValidator   Validator
	Modify         Executor
	Validator      Validator
	Validators     []Validator
	ValidatorRetry RetryOptions
	Substep        *Step
	Substeps       []*Step
	SubstepRetry   RetryOptions
	// A simple way to invert the meaning of the Validator.  Validators
	// are encouraged to provide more specific negative testing behaviors,
	// but this serves for simpler testing.  If set, it inverts the
	// response from the call sent to Retry, so it can be used to wait
	// until an error is returned (but there is no control on which kind
	// of error that will be)
	ExpectError bool
	// TODO: ExpectIs, ExpectAs; use errors.Is, errors.As against a list of expected errors?
	SkipWhen bool
}

func (s Step) Logf(format string, v ...interface{}) {
	if s.IsVerbose() {
		left := strings.Repeat(" ", s.Level)
		log.Printf(left+format, v...)
	}
}

func (s Step) IsVerbose() bool {
	return s.Verbose || os.Getenv(EnvFrame2Verbose) != ""
}

// type Validate struct {
// 	Validator
// 	// Every Validator runs inside a Retry.  If no options are given,
// 	// the default RetryOptions are used (ie, single run of Fn, with either
// 	// failed check or error causing the step to fail)
// 	RetryOptions
// }
//
// func (v Validate) GetRetryOptions() RetryOptions {
// 	return v.RetryOptions
// }

type Validator interface {
	Validate() error
}

// TODO create ValidatorList, with Validator + RetryOptions

type Execute struct {
}

type Executor interface {
	Execute() error
}

type TearDowner interface {
	Teardown() Executor
}
