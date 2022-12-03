package frame2

// The check return value indicates whether the function succeeded
// in its main purpose.
// The err indicates an unexpected error
type CheckedRetryFunction func() (check bool, err error)

type RetryResult struct {
	err   error
	check bool
}

type Retry struct {
	Fn      CheckedRetryFunction // The thing to be retried
	Options RetryOptions
}

func (r Retry) Run() []RetryResult {
	return nil
}

// Runs the retry in parallel; returns a function
// that will wait and return the results only
// when it finished (wait).
// TODO perhaps give it a context, too
func (r Retry) ParallelRun() func() []RetryResult {
	return nil
}

// The maximum number of retries will be:
// max (allow, ignore) + Ensure + Retries
// but it will often be less.  the minimum:
// 0 (for allow/ignore that were immediatelly matched) + Ensure (where no retry past Ensure was required)
type RetryOptions struct {
	Allow   int // for initial failures
	Ignore  int // initial successes
	Ensure  int // last n tries are successful
	Retries int // after initial allow/ignore; can be zero.  How does it play with Ensure?  Negative means forever?
	// regardless of allow/ignore/maxRetries, the two below will cause
	// a return with an error
	FatalOnCheck bool
	FatalOnError bool
	// If both of these are false, either an error or a false check will
	// cause the retry to report a failure, after the retries are exhausted
	IgnoreCheck bool // At the end, is the bool still false?
	IgnoreError bool // At the end is th error still being returned?
	Context     bool // aggregate timed with number of tries; either or both can be used
	Verbose     bool // Log every error?
}
