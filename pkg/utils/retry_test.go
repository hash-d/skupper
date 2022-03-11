package utils

import (
	"fmt"
	"testing"
	"time"
)

//     f() return:    Retry()
//
// n   ok     err     maxRetries  return (error)                     today           original
// --  ----- -----    ----------- --------------                     ------          --------
// 1   true,   nil    no          nil                                ✓ nil           nil
// 2   true,  !nil    no          retry?                             ? nil           err from f()
// 3   false,  nil    no          retry?                             ? retry         retry
// 4   false, !nil    no          retry                              ✓ retry         err from f()
//
// 5   true,   nil    yes         nil                                ✓ nil           nil
// 6   true,  !nil    yes         err from f()?                      ? err from f()  err from f()
// 7   false,  nil    yes         new err ("Max retries reached")?   x retry (*)     RetryError
// 8   false, !nil    yes         err from f()                       ✓ err from f()  err from f()
//
// (*) possibly an infinite loop
//
//
// So, what's ok and err, and how should they be treated?  If they go together
// (ok == (err != nil)), everything works fine, but if they differ, things get
// weird
//
// Ok may mean two things:
//
// 1 - keep trying.  In that case, we'd be returning 'err' from f() only to
//     pass it along.
//
// 2 - stop and fail.  If that was so, we should remove the check by err and
//     replace it by a check with ok
//
// In either case, we may end up on the weird situation where we return before
// maxRetries with err == ""
//
// To me, it makes more sense if ok means 'keep trying if error, return nil if
// no error (and till maxRetries).  That way, 'ok' could be used to
// differentiate between fatal and non-fatal errors (for example, if a
// connection error, return false; if it connected, but did not have the
// expected response, then keep trying.  The caller then would be able to
// differentiate them by the returned error.
//
// The way the function was originally written reads as this:
//
// - If function produces an error, fail immediatelly with that error
// - Else, if ok is true, return nil and succeed
// - Otherwise, retry
//
// So, it is not retry on error, it is retry until ok, and stop on error
//
// Documentation update proposals:
// - what, exactly, is ok, and how it interacts with err
// - make it clear that maxRetries = 1 means that the function will run at most
//   twice (the original run + one retry at maximum)

type RetryTestItem struct {
	// These two configure what the f() function will respond
	ok  bool
	err error
	// This configures Retry itself
	maxRetries int
	// And those are what we're expecting the actual result to look like
	expectedRetries  int // rename this to expectedTries?  That would be first try + retries
	expectedResponse error

	// perhaps change this to okOn and nilOn
	succeedOn int
}

func TestRetry(t *testing.T) {

	testTable := []RetryTestItem{
		{ // #1
			ok:               true,
			err:              nil,
			maxRetries:       3,
			expectedRetries:  1,
			expectedResponse: nil,
		}, {
			ok:               true,
			err:              nil,
			maxRetries:       -1,
			expectedRetries:  0,
			expectedResponse: fmt.Errorf("maxRetries (%d) should be > 0", -1),
		}, {
			ok:               true,
			err:              nil,
			maxRetries:       0,
			expectedRetries:  0,
			expectedResponse: fmt.Errorf("maxRetries (%d) should be > 0", 0),
		}, { // #4, #8
			ok:               false,
			err:              fmt.Errorf("app error"),
			maxRetries:       3,
			expectedRetries:  3,
			expectedResponse: fmt.Errorf("app error"),
		}, { // #2, ~#6~
			ok:               true,
			err:              fmt.Errorf("app error"),
			maxRetries:       3,
			expectedRetries:  3,
			expectedResponse: fmt.Errorf("app error"),
		}, { // #1, #5
			ok:               false,
			err:              nil,
			maxRetries:       3,
			expectedRetries:  2,
			expectedResponse: nil,
			succeedOn:        2,
		}, { // #1, #5; extreme
			ok:               false,
			err:              nil,
			maxRetries:       3,
			expectedRetries:  3,
			expectedResponse: nil,
			succeedOn:        3,
		}, { // #3, #7
			ok:               false,
			err:              nil,
			maxRetries:       3,
			expectedRetries:  3,
			expectedResponse: nil, // this will loop forever
		},
	}

	for _, item := range testTable {
		name := fmt.Sprintf("ok:%v err:%v retries:%v maxRetries:%v succeedOn:%v", item.ok, item.err, item.expectedRetries, item.maxRetries, item.succeedOn)

		var currentTry int
		t.Run(name, func(t *testing.T) {

			retryErr := Retry(time.Second, item.maxRetries, func() (bool, error) {
				currentTry++
				if currentTry > item.maxRetries+1 {
					// This is a protection for infinite loops
					t.Fatalf("Retry %v > maxRetries %v + 1", currentTry, item.maxRetries)
				}
				if currentTry == item.succeedOn {
					return true, nil
				}
				return item.ok, item.err
			})

			if item.expectedResponse != nil {
				if retryErr != nil {
					if retryErr.Error() != item.expectedResponse.Error() {
						t.Error(retryErr)
					}
				} else {
					t.Error(retryErr)
				}
			} else {
				if retryErr != nil {
					t.Error(retryErr)
				}
			}

			if currentTry != item.expectedRetries {
				t.Errorf("%v != %v", currentTry, item.expectedRetries)
			}

		})
	}

}
