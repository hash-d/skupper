//go:build policy
// +build policy

// hello_policy implements a series of tests on the Policy engine, based on the pre-existing
// hello_world example test.
//
// Each policy piece has its own file.  On it, we define both the
// piece-specific tests _and_ the piece-specific infra.
//
// There is a single root test on the whole package, located on main_test.go and named
// TestPolicies.  As the Policy CRD and CRs are cluster-wide, all tests need to be
// run in serial, and TestPolicies is responsible for that.  For this reason, the
// individual tests' functions are named testXxx (lower first character) so they're
// not called from `go test` directly.
//
// The 'piece-specific' infra is a set of functions that mostly return a cli.TestScenario
// (or a list thereof).  These are called from from the actual tests, to provide the
// pieces that are then combined into an actual test.
//
// More than saving keystrokes, the idea on having these helpers is to:
//
// - Make the testing consistent.  As it is based on hello_world, each piece is a
//   copy of the step done on hello_world, as much as possible.  The functions help
//   avoiding the tests from deviating from that.  Also, if that standard changes,
//   it needs changed in a single place
//
// - The tests become more readable.  Instead of a long structure with details on
//   what is being done, the tests have a function call whose name and godoc indicate
//   what is doing.
//
// Each of these functions are placed on the same file that holds the test that
// is more closely related to them.  For example, the checking for link being
// (un)able to create or being destroyed is defined on functions on
// link_test.go
//
// These functions will take a cluster context and an optional name prefix.
// They will return a cli.TestScenario with the intended objective on the
// requested cluster, and the name of the scenario will receive the prefix, if
// any given.  A use of that prefix would be, for example, to clarify that
// what's being checked is a 'side-effect' (eg when a link drops in a cluster
// because the policy was removed on the other cluster)
//
// The runner is structured as follows:
//
//     policyTestRunner (keepPolicies bool, background policies, contextMap)
//       []policyTestCase (just a name and a set of steps)
//         []policyTestStep (a name and all the actual configuration of the test: what to execute and how)
//             preHook
//             policies
//             GET checks
//             cliScenarios  ([]cli.TestScenario)
//
//             parallel (decide between cli.RunScenarios and cli.RunScenariosParallel)
//             skipFunction
//             sleep (post execution)
//
//
//     cli.TestScenario (name and list of tasks)
//       []cli.SkupperTask (which cluster to run, and a list of commands)
//         []cli.SkupperCommandTester (an interface; each item represent an individual call to the skupper binary)
//
// As described above, the policyTestRunner is a single structure per test,
// which contain configurations that are valid for multiple test cases.  A
// policyTestCase is a purely organizational structure: it identifies an
// individual test case within the test.  It has only a name and a list of
// policyTestStep.  policyTestStep is where the tests are really defined.
// Check its godoc for details.
//
// Writing tests
//
// When writing tests, keep in mind that the policies are cluster-wide.  So, while the runner's policyTestStep
// takes care to install the policy on the cluster associated with the respective context, that cluster may
// be the same or a different cluster.
//
// If you create policies that affect namespaces on both clusters, you'll have to either check
// base.MultipleClusters() or create the policies in such a way that the test works on both situations.
package hello_policy

// TODO
//
// Pre-merge (priority on top)
// - Document
//   - The whole (how do things fit together)
//   - The items (each test file)
//   - Rationale for individual test cases/steps
// - Stop at start if CRD already present (avoid changing pre-existing policies); update CI
// - Review Fernando's PR
//   - Better CRD removal at the end; check that two contexts point to the same cluster
// - Check TODO across the code
// - Check 'ExpectAuthError': change name to 'Expect no service'?
// - Add GETs, make test overal less flaky
// - Reorganize test calling from main_test
// - Ensure it works with upstream CI (especially host checking)
// - Check on status for multicluster checking
// - Re-implement hello_world using runner, composing functions
// - Dump, capture debug info on errors
// - List points where I could get help for better solutions (../../../../.../crd)
//
// Post-merge (priority on top)
// - Confirm 'not-bound' checks are really checking services for not being bound
// - Check for tests that need better finish
//   - AllowedOutgoingLinksHostnames
//     - Cross testing (claim on router and vice versa)
//     - full setup checking (create service and expose; check they appear/disappear; perhaps even curl the service)
//     - different removals and reinstates of policy (actual removal, changed namespace list)
// - Define how specific-issue (reproducer) tests are going to be handled
// - Non-admin skupper init
// - Non-admin user (or: use admin only for CRD/CR, init)
// - Review test structure.  In special repeated items (test_name#01)
// - Check test coverage (specific image and all)
// - Additional tests: gateway, annotation, upgrade, console
// - Operator + config map
//
