//go:build meta_test
// +build meta_test

package frame2

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

var funcError = errors.New("FuncError")

type testChecks struct {
	input  []error
	result error
}

type test struct {
	config RetryOptions
	checks []testChecks
}

func TestRetry(t *testing.T) {
	table := []test{
		{
			// No retries, so the second item in the input should
			// never be returned
			config: RetryOptions{Interval: time.Millisecond},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Retries:  1,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   1,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Allow:    1,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Ignore:   2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError},
					result: funcError,
				}, {
					input:  []error{nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Allow:    2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Retries:  4,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, funcError, nil, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   2,
				Allow:    2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil, nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, nil, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, nil, funcError, funcError, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Allow:    2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   2,
				Ensure:   2,
				Allow:    2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, nil, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, nil, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, funcError},
					result: funcError,
				},
			},
		},
	}

	for i, item := range table {
		t.Run(fmt.Sprintf("item-%d", i), func(t *testing.T) {
			for j, c := range item.checks {
				t.Run(fmt.Sprintf("check-%d", j), func(t *testing.T) {
					var n int
					_, err := Retry{
						Options: item.config,
						Fn: func() error {
							if n >= len(c.input) {
								t.Logf("Test failed: %+v", item)
								t.Fatalf("tried to access input #%d", n)
							}
							ret := c.input[n]
							n++
							return ret
						},
					}.Run()
					t.Logf("Got response %v", err)
					if !errors.Is(err, c.result) {
						t.Logf("Test failed: %+v", item)
						t.Errorf("%v != %v", err, c.result)
					}
					if n != len(c.input) {
						t.Logf("Test failed: %+v", item)
						t.Errorf("used %v items from total %v in input", n, len(c.input))
					}
				})
			}
		})

	}
}
