package frame2

import (
	"fmt"
	"log"
	"time"
)

var (
	lastSuccessErrorStr = "cannot ensure requested number of successes.  last error: %w"
)

type RetryFunction func() (err error)

type Retry struct {
	Fn      RetryFunction // The thing to be retried
	Options RetryOptions
}

// The maximum number of retries will be:
// max (allow, ignore) + Ensure + Retries
// but it will often be less.  the minimum:
// 0 (for allow/ignore that were immediatelly matched) + Ensure (where no retry past Ensure was required)
//
// Allow accounts for instabilities (for example, a service with two providers might return a mix of
// success and failures while the two providers stabilize).  The last successful results (if not ignored)
// will count to Ensure.
//
// Once past the Allow phase, any errors reset the success count.  Retry will kick in if configured, and
// if the remaining tries fit into the Ensure target.
//
// The ignore counts from the first success in the last success streak from the Allow phase, or from
// the start of the retry phase, if no allow configured or no success in that phase
//
type RetryOptions struct {
	Allow    int           // for initial failures
	Ignore   int           // initial successes
	Ensure   int           // last n tries are successful
	Retries  int           // after initial allow/ignore; can be zero.  How does it play with Ensure?  Negative means forever?
	Interval time.Duration // if not given, the default is 1s
	// regardless of allow/ignore/maxRetries, the two below will cause
	// a return with an error
	//	FatalOnCheck bool
	//	FatalOnError bool
	// If both of these are false, either an error or a false check will
	// cause the retry to report a failure, after the retries are exhausted
	//	IgnoreCheck bool // At the end, is the bool still false?
	//	IgnoreError bool // At the end is th error still being returned?

	//	Context     bool // aggregate timed with number of tries; either or both can be used
	//	Verbose     bool // Log every error?
}

func (r Retry) Run() ([]error, error) {
	interval := time.Duration(r.Options.Interval)
	if interval == 0 {
		interval = time.Second
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()

	results := []error{}

	var totalTries int
	var consecutiveSuccess int
	var ignoredSuccess int
	var retries int
	var ensure = r.Options.Ensure
	if ensure == 0 {
		ensure = 1
	}
	//var anySuccess bool

	//	var lastError error

	for {
		totalTries++
		err := r.Fn()
		results = append(results, err)
		if err == nil {
			//anySuccess = true
			// Are we counting this as a success?
			if ignoredSuccess >= r.Options.Ignore || totalTries > r.Options.Ignore {
				consecutiveSuccess++
			} else {
				ignoredSuccess++
			}
			// At least one success that counts + whatever Ensure wants, and we're good
			if consecutiveSuccess >= ensure {
				return results, nil
			}
			continue
			// It's a success, but not enough; can we still make it?
			//			if totalTries-ignoredTries > r.Options.Allow+r.Options.Retries {
			//				err = fmt.Errorf(lastSuccessErrorStr, lastError)
			//				return results, err
			//			} else {
			//				continue
			//			}
		}
		// This try failed, and we ran out of retries.  Note retries only count after Allow expires
		if totalTries > r.Options.Allow && retries >= r.Options.Retries {
			return results, fmt.Errorf("max retry attempts reached: %w", err)
		}
		remainingRetries := r.Options.Retries - retries // at this point
		if remainingRetries < 0 {
			remainingRetries = 0
		}
		//		if totalTries > r.Options.Allow && ensure > totalTries-r.Options.Allow-remainingRetries {
		if totalTries > r.Options.Allow && false { //&& remainingRetries < 0 {
			// All hope is lost, we cannot comply with Ensure anymore, even considering
			// allow and retries
			log.Printf("remaining: %d", remainingRetries)
			log.Printf("total:     %d", totalTries)
			log.Printf("ensure:    %d", ensure)
			return results, fmt.Errorf(lastSuccessErrorStr, err)
		}
		//		lastError = err
		consecutiveSuccess = 0
		ignoredSuccess = 0
		// Any errors past the initial allow number + retries are fatal, but we
		// do not consider any ignored tries
		//		if totalTries > r.Options.Allow+r.Options.Retries {
		//			return results, fmt.Errorf("ASDF: %w", err)
		//		}
		// If I got down here and it's past Allow time, the next run will be a retry
		if totalTries > r.Options.Allow {
			retries++
		}
		//		if totalTries > r.Options.Ignore+r.Options.Allow+r.Options.Ensure+r.Options.Retries {
		//			return results, err
		//		}
		<-tick.C
	}
	//return failures, nil
}

// Runs the retry in parallel; returns a function
// that will wait and return the results only
// when it finished (wait).
// TODO perhaps give it a context, too
func (r Retry) ParallelRun() func() []error {
	return nil
}
