package frame2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type RetryFunction func() (err error)

type Retry struct {
	Fn      RetryFunction // The thing to be retried
	Options RetryOptions
}

// Allow accounts for instabilities (for example, a service load balanced on
// two providers might return a mix of successes and failures while the two
// providers stabilize).  The last success streak in this phase will count to
// Ensure.
//
// Once past the Allow phase, any errors will cause a failure, unless there
// are Retries available
//
// Even successes may require additional runs.  There are two cases here:
//
//   - If Ensure is set, the test will keep trying on success until the required
//     number of successes are met
//   - If Ignore was set, that number of successes will be ignored on the count
//     to Ensure, possibly requiring additional runs until the Ensure target is met
//
// These, however, do not count as Retries.  i. e, Retries are only those
// additional runs that were caused by a failure.
//
// The ignore counts from the first success in the last success streak from the
// Allow phase, or from the start of the retry phase (if no Allow configured or
// no success in that phase)
type RetryOptions struct {
	Allow    int           // for initial failures
	Ignore   int           // initial successes
	Ensure   int           // last n tries are successful.  Minimum 1
	Retries  int           // after Allow phase
	Interval time.Duration // if not given, the default is 1s
	Quiet    bool          // if true, no attempt logs
	//	Context     bool // aggregate timed with number of tries; either or both can be used
	//	Verbose     bool // Log every error?

	Min int // TODO Run as normal, but delay report until that number of tries have been done
	// This can be used to generate stats from the results

	Rate float32 // TODO: "Ensure" will not be 100%, but based on this rate.  So, if Ensure is 100
	// And Rate 50%, success will be achieved with at least 50 success in the past
	// 100 executions

	KeepTrying bool
	Ctx        context.Context
	Timeout    time.Duration
}

// Checks whether any fields on the struct have been set
func (r RetryOptions) IsEmpty() bool {
	if r.Allow != 0 ||
		r.Ignore != 0 ||
		r.Ensure != 0 ||
		r.Retries != 0 ||
		r.Interval != 0 ||
		r.Quiet != false ||
		r.Min != 0 ||
		r.Rate != 0 ||
		r.KeepTrying != false ||
		r.Ctx != nil ||
		r.Timeout != 0 {
		return false
	}

	return true
}

func (r Retry) Run() ([]error, error) {
	interval := r.Options.Interval
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
	if ensure < 1 {
		ensure = 1
	}
	ctx := r.Options.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	var cn context.CancelFunc
	if r.Options.Timeout > 0 {
		ctx, cn = context.WithTimeout(ctx, r.Options.Timeout)
		defer cn()
	}
	for {
		// Before any tries, check the context
		err := ctx.Err()
		if err != nil {
			if !r.Options.Quiet {
				log.Printf("retry cancelled: %v", err)
			}
			return results, err
		}

		totalTries++
		err = r.Fn()
		results = append(results, err)
		var contextInfo string
		if dl, ok := ctx.Deadline(); ok {
			contextInfo = fmt.Sprintf(" [timeout in %v]", dl.Sub(time.Now()))
		}
		if err == nil {
			// Are we counting this as a success?
			if ignoredSuccess >= r.Options.Ignore || totalTries > r.Options.Ignore {
				consecutiveSuccess++
			} else {
				ignoredSuccess++
			}
			// Are we good?
			if consecutiveSuccess >= ensure {
				if totalTries > 1 {
					log.Printf("Success on attempt %v", totalTries)
				}
				return results, nil
			}
			// It's a success, but not enough; we'll try again
			if !r.Options.Quiet {
				var info = []string{}
				if r.Options.Ensure > 1 {
					info = append(info, fmt.Sprintf("%d/%d consecutive successes", consecutiveSuccess, r.Options.Ensure))
				}
				if r.Options.Ignore > 0 {
					info = append(info, fmt.Sprintf("%d/%d ignored", ignoredSuccess, r.Options.Ignore))
				}
				info = append(info, contextInfo)

				msg := fmt.Sprintf("Attempt %d succeeded; ", totalTries)

				msg = msg + strings.Join(info, ", ")

				log.Printf(msg)
			}
			<-tick.C
			continue
		}
		if !r.Options.KeepTrying {
			// This try failed, and we ran out of retries.  Note retries only count after Allow expires
			if totalTries > r.Options.Allow && retries >= r.Options.Retries {
				if r.Options.Retries > 1 {
					return results, fmt.Errorf("max retry attempts reached: %w", err)
				} else {
					return results, err
				}

			}
		}
		consecutiveSuccess = 0
		ignoredSuccess = 0
		// If I got down here and it's past Allow time, the next run will be a retry
		if totalTries > r.Options.Allow {
			retries++
		}
		if !r.Options.Quiet {
			msg := fmt.Sprintf(
				"Attempt %d failed (allow %d first + %d/%d retries used)",
				totalTries, r.Options.Allow, retries, r.Options.Retries,
			)
			if r.Options.KeepTrying {
				msg += " [keep trying]"
			}
			msg += contextInfo
			log.Print(msg)
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
