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
	Level     int
	// Whether the step should always print logs
	// Even if false, logs will be done if SKUPPER_TEST_FRAME2_VERBOSE
	Verbose        bool
	PreValidator   Validator
	Modify         Executor
	Validator      Validator
	ValidatorRetry RetryOptions
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
