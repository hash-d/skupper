package utils

import (
	"fmt"
	"testing"
	"time"
)

//     f() return:    Retry()
//
// n   ok     err     maxRetries  return (error)                     today
// --  ----- -----    ----------- --------------                     ------
// 1   true,   nil    no          nil                                ✓ nil
// 2   true,  !nil    no          retry?                             ? nil
// 3   false,  nil    no          retry?                             ? retry
// 4   false, !nil    no          retry                              ✓ retry
//
// 5   true,   nil    yes         nil                                ✓ nil
// 6   true,  !nil    yes         err from f()?                      ? err from f()
// 7   false,  nil    yes         new err ("Max retries reached")?   x retry          <--- possibly an infinite loop
// 8   false, !nil    yes         err from f()                       ✓ err from f()
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

type RetryTestItem struct {
	ok               bool
	err              error
	maxRetries       int
	expectedRetries  int
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

		var currentRetry int
		t.Run(name, func(t *testing.T) {

			retryErr := Retry(time.Second, item.maxRetries, func() (bool, error) {
				currentRetry++
				if currentRetry > item.maxRetries {
					// This is a protection for infinite loops
					return false, fmt.Errorf("Retry %v > maxRetries %v", currentRetry, item.maxRetries)
				}
				if currentRetry == item.succeedOn {
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

			if currentRetry != item.expectedRetries {
				t.Errorf("%v != %v", currentRetry, item.expectedRetries)
			}

		})
	}

}
