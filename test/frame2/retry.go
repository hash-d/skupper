package frame2

import (
	"fmt"
	"time"
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

	// We have to have at least one success
	var ensure = r.Options.Ensure
	if ensure == 0 {
		ensure = 1
	}
	for {
		totalTries++
		err := r.Fn()
		results = append(results, err)
		if err == nil {
			// Are we counting this as a success?
			if ignoredSuccess >= r.Options.Ignore || totalTries > r.Options.Ignore {
				consecutiveSuccess++
			} else {
				ignoredSuccess++
			}
			// Are we good?
			if consecutiveSuccess >= ensure {
				return results, nil
			}
			// It's a success, but not enough; we'll try again
			continue
		}
		// This try failed, and we ran out of retries.  Note retries only count after Allow expires
		if totalTries > r.Options.Allow && retries >= r.Options.Retries {
			return results, fmt.Errorf("max retry attempts reached: %w", err)
		}
		consecutiveSuccess = 0
		ignoredSuccess = 0
		// If I got down here and it's past Allow time, the next run will be a retry
		if totalTries > r.Options.Allow {
			retries++
		}
		<-tick.C
	}
}

// Runs the retry in parallel; returns a function
// that will wait and return the results only
// when it finished (wait).
// TODO perhaps give it a context, too
func (r Retry) ParallelRun() func() []error {
	return nil
}
