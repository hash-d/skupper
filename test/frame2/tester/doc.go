// The files in this package contains two types of strucs:
//
// - Full testers, that will execute an action and inspect every detail
//   of its execution that's specified
// - Validators used by these testers.  They can be reused when calling
//   the base command (for quicker execution), but one still wants to
//   run some of the validations.
//
//   These are used by waiters, as well.
//
//   Example:
//
//   // TODO: change the examples for something that actually does changes
//
//   exec.CliLinkStatus (the base executor)
//   waiter.CliLinkStatus (executes, then tests only enough to ensure we can go to the next step of a test)
//   tester.CliLinkStatus (provides a full test)
package tester
