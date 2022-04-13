package hello_policy

import (
	"strconv"
	"testing"

	"github.com/skupperproject/skupper/client"
	skupperv1 "github.com/skupperproject/skupper/pkg/apis/skupper/v1alpha1"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

var (
	// these are final, do not change them.  They're used with
	// a boolean pointer to allow true/false/undefined
	_true  = true
	_false = false
	// TODO: is there a better way to do this?
)

// Each policy piece has its own file.  On it, we define both the
// piece-specific tests _and_ the piece-specific infra.
//
// For example, the checking for link being (un)able to create or being
// destroyed is defined on functions on link_test.go
//
// These functions will take a cluster context and an optional name prefix.  It
// will return a slice of cli.TestScenario with the intended objective on the
// requested cluster, and the names of the individual scenarios will receive
// the prefix, if any given.  A use of that prefix would be, for example, to
// clarify that what's being checked is a 'side-effect' (eg when a link drops
// in a cluster because the policy was removed on the other cluster)
//
// policyTestRunner
//   []policyTestCase
//     []policyTestStep
//         policies
//         cli commands
//         GET checks

// Runs each policyTestCase in turn
//
// By default, all policies are removed between the tests cases, but that can be
// controlled with keepPolicies
type policyTestRunner struct {
	scenarios    []policyTestCase
	keepPolicies bool
}

// Runs each test case in turn
func (r policyTestRunner) run(t *testing.T, pub, prv *base.ClusterContext) {

	for _, testCase := range r.scenarios {
		if !r.keepPolicies {
			removePolicies(t, pub)
			removePolicies(t, prv)
		}
		if base.IsTestInterrupted() {
			break
		}
		testCase.run(t, pub, prv)
	}
	base.StopIfInterrupted(t)
}

// A named slice, with methods to run each step
type policyTestCase struct {
	name  string
	steps []policyTestStep
}

// Runs the individual steps in a test case.  The test case is an individual
// Go test
func (c policyTestCase) run(t *testing.T, pub, prv *base.ClusterContext) {

	t.Run(
		c.name,
		func(t *testing.T) {
			for _, step := range c.steps {

				step.run(t, pub, prv)
				if base.IsTestInterrupted() {
					break
				}
			}
			base.StopIfInterrupted(t)
		})
}

// Configures a step on the policy test runner, which allows for setting
// policies on the two clusters, run a set of cli commands and then perform
// some checks using the `get` command.
//
// ATTENTION to how the policy lists (pubPolicy, prvPolicy) work:
// - Each item on the list will generate a policy named pub/prv-policy-i,
//   based on their position on the list (i is an index)
// - Every time a list is defined, each of its items will be either updated
//   or created
//
// So, if the previous step defined two public policies, and the current step...
//
// - defines none: nothing is changed; the two policies stay in place
// - defines only one: the first policy is updated; the second one is not touched
// - defines two policies: both are updated
// - defined three policies: the first two are updated; the third one created
//
// You may use this behavior on your tests, by placing changing policies at the
// start of the list, and never-changing at the end, so your updates will simply
// have the first one or two policies listed.  However, be careful, it is easy
// to overlook this behavior causing weird test errors.
type policyTestStep struct {
	name        string
	pubPolicy   []skupperv1.SkupperClusterPolicySpec // ATTENTION to usage; see doc
	prvPolicy   []skupperv1.SkupperClusterPolicySpec
	commands    []cli.TestScenario
	pubGetCheck policyGetCheck
	prvGetCheck policyGetCheck
	parallel    bool // This will run the cli commands in parallel
}

// Runs the TestStep as an individual Go Test
func (s policyTestStep) run(t *testing.T, pub, prv *base.ClusterContext) {
	t.Run(
		s.name,
		func(t *testing.T) {
			s.applyPolicies(t, pub, prv)
			s.runCommands(t, pub, prv)
			s.runChecks(t, pub, prv)
		})
}

// Apply all policies, on pub and prv
//
// See policyTestStep documentation for behavior
func (s policyTestStep) applyPolicies(t *testing.T, pub, prv *base.ClusterContext) {

	if len(s.pubPolicy)+len(s.prvPolicy) > 0 {
		t.Run(
			"policy-setup",
			func(t *testing.T) {
				for i, policy := range s.pubPolicy {
					i := strconv.Itoa(i)
					err := applyPolicy(t, "pub-policy-"+i, policy, pub)
					if err != nil {
						t.Fatalf("Failed to apply policy: %v", err)
					}
				}
				for i, policy := range s.prvPolicy {
					i := strconv.Itoa(i)
					err := applyPolicy(t, "prv-policy-"+i, policy, prv)
					if err != nil {
						t.Fatalf("Failed to apply policy: %v", err)
					}
				}

			})
	}
}

// Runs a policy check using `get`.  It receives a *testing.T, so it does not return;
// it marks the test as failed if the check fails.
func getChecks(t *testing.T, getCheck policyGetCheck, c *client.PolicyAPIClient) {
	ok, err := getCheck.check(c)
	if err != nil {
		t.Errorf("GET check failed with error: %v", err)
		return
	}

	if !ok {
		t.Errorf("GET check failed (check: %v)", getCheck)
	}
}

// Runs the configured GET checks
func (s policyTestStep) runChecks(t *testing.T, pub, prv *base.ClusterContext) {
	pubPolicyClient := client.NewPolicyValidatorAPI(pub.VanClient)
	prvPolicyClient := client.NewPolicyValidatorAPI(prv.VanClient)

	getChecks(t, s.pubGetCheck, pubPolicyClient)
	getChecks(t, s.prvGetCheck, prvPolicyClient)
}

// Run the commands part of the policyTestStep
func (s policyTestStep) runCommands(t *testing.T, pub, prv *base.ClusterContext) {
	if s.parallel {
		cli.RunScenariosParallel(t, s.commands)
	} else {
		cli.RunScenarios(t, s.commands)
	}
}
