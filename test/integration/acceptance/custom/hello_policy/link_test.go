package hello_policy

import (
	"testing"

	skupperv1 "github.com/skupperproject/skupper/pkg/apis/skupper/v1alpha1"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
	"github.com/skupperproject/skupper/test/utils/skupper/cli/link"
)

// Uses the named token to create a link from ctx
//
// Returns a scenario with a single link.CreateTester
//
// It runs no test task, as the link may have been created but as inactive
func createLinkTestScenario(ctx *base.ClusterContext, prefix, name string) (scenario cli.TestScenario) {

	scenario = cli.TestScenario{
		Name: prefixName(prefix, "connect-sites"),
		Tasks: []cli.SkupperTask{
			{
				Ctx: ctx, Commands: []cli.SkupperCommandTester{
					&link.CreateTester{
						TokenFile: "./tmp/" + name + ".token.yaml",
						Name:      name,
						Cost:      1,
					},
				},
			},
		},
	}

	return
}

// Produces a TestScenario named link-is-up/down, and checks accordingly
func linkStatusTestScenario(ctx *base.ClusterContext, prefix, name string, up bool) (scenario cli.TestScenario) {
	var statusStr string

	if up {
		statusStr = "up"
	} else {
		statusStr = "down"
	}

	scenario = cli.TestScenario{
		Name: prefixName(prefix, "link-is-"+statusStr),
		Tasks: []cli.SkupperTask{
			{
				Ctx: ctx,
				Commands: []cli.SkupperCommandTester{
					&link.StatusTester{
						Name:   name,
						Active: up,
					},
				},
			},
		},
	}

	return
}

// Returns a TestScenario that calls skupper link delete on the named link.
//
// The scenario will be called remove-link
func linkDeleteTestScenario(ctx *base.ClusterContext, prefix, name string) (scenario cli.TestScenario) {
	scenario = cli.TestScenario{
		Name: prefixName(prefix, "remove-link"),
		Tasks: []cli.SkupperTask{
			{
				Ctx: ctx,
				Commands: []cli.SkupperCommandTester{
					&link.DeleteTester{
						Name: name,
					},
				},
			},
		},
	}
	return
}

// Returns a TestScenario that confirms two sites are connected.  The check is
// done on both sides.
//
// The scenario name is validate-sites-connected, and the configuration should
// match the main hello_world test
func sitesConnectedTestScenario(pub *base.ClusterContext, prv *base.ClusterContext, prefix, linkName string) (scenario cli.TestScenario) {

	scenario = cli.TestScenario{
		Name: prefixName(prefix, "validate-sites-connected"),
		Tasks: []cli.SkupperTask{
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper status - verify sites are connected
				&cli.StatusTester{
					RouterMode:          "interior",
					ConnectedSites:      1,
					ConsoleEnabled:      true,
					ConsoleAuthInternal: true,
					PolicyEnabled:       true,
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper status - verify sites are connected
				&cli.StatusTester{
					RouterMode:     "edge",
					SiteName:       "private",
					ConnectedSites: 1,
					PolicyEnabled:  true,
				},
				// skupper link status - testing all links
				&link.StatusTester{
					Name:   linkName,
					Active: true,
				},
				// skupper link status - now using link name and a 10 secs wait
				&link.StatusTester{
					Name:   linkName,
					Active: true,
					Wait:   10,
				},
			}},
		},
	}
	return
}

// Return a SkupperClusterPolicySpec that (dis)allows incomingLinks on the
// given namespace.
func allowIncomingLinkPolicy(namespace string, allow bool) (policySpec skupperv1.SkupperClusterPolicySpec) {
	policySpec = skupperv1.SkupperClusterPolicySpec{
		Namespaces:         []string{namespace},
		AllowIncomingLinks: allow,
	}

	return
}

// Return a SkupperClusterPolicySpec that allows outgoing links to the given
// hostnames (a string list, following the policy's specs) on the given
// namespace.
func allowedOutgoingLinksHostnamesPolicy(namespace string, hostnames []string) (policySpec skupperv1.SkupperClusterPolicySpec) {
	policySpec = skupperv1.SkupperClusterPolicySpec{
		Namespaces:                    []string{namespace},
		AllowedOutgoingLinksHostnames: hostnames,
	}

	return
}

func testLinkPolicy(t *testing.T, pub, prv *base.ClusterContext) {

	testTable := []policyTestCase{
		{
			name: "init",
			steps: []policyTestStep{
				{
					name:     "execute",
					parallel: true,
					commands: []cli.TestScenario{
						skupperInitInteriorTestScenario(pub, "", true),
						skupperInitEdgeTestScenario(prv, "", true),
					},
				},
			},
		},
		{
			name: "empty-policy-fails-token-creation",
			steps: []policyTestStep{
				{
					name: "execute",
					commands: []cli.TestScenario{
						createTokenPolicyScenario(pub, "", "./tmp", "fail", false),
					},
					pubGetCheck: policyGetCheck{
						checkUndefinedAs: &_false,
					},
				},
			},
		}, {
			name: "allowing-policy-allows-creation",
			steps: []policyTestStep{
				{
					name: "execute",
					pubPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowIncomingLinkPolicy(pub.Namespace, true),
					},
					prvPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowedOutgoingLinksHostnamesPolicy(prv.Namespace, []string{"*"}),
					},
					commands: []cli.TestScenario{
						createTokenPolicyScenario(pub, "", "./tmp", "works", true),
						createLinkTestScenario(prv, "", "works"),
						linkStatusTestScenario(prv, "", "works", true),
						sitesConnectedTestScenario(pub, prv, "", "works"),
					},
					pubGetCheck: policyGetCheck{
						allowIncoming:    &_true,
						checkUndefinedAs: &_false,
					},
					prvGetCheck: policyGetCheck{
						allowIncoming: &_false,
					},
				}, {
					name: "remove",
					pubPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowIncomingLinkPolicy(pub.Namespace, false),
					},
					commands: []cli.TestScenario{
						linkStatusTestScenario(prv, "", "works", false),
					},
					pubGetCheck: policyGetCheck{
						allowIncoming: &_false,
					},
					prvGetCheck: policyGetCheck{
						allowIncoming: &_false,
					},
				}, {
					name: "re-allow",
					pubPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowIncomingLinkPolicy(pub.Namespace, true),
					},
					commands: []cli.TestScenario{
						linkStatusTestScenario(prv, "again", "works", true),
						sitesConnectedTestScenario(pub, prv, "", "works"),
						linkDeleteTestScenario(prv, "", "works"),
					},
					pubGetCheck: policyGetCheck{
						allowIncoming:    &_true,
						checkUndefinedAs: &_false,
					},
					prvGetCheck: policyGetCheck{
						allowIncoming: &_false,
					},
				},
			},
		}, {
			name: "previously-created-token",
			steps: []policyTestStep{
				{
					name: "prepare",
					pubPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowIncomingLinkPolicy(pub.Namespace, true),
					},
					prvPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowedOutgoingLinksHostnamesPolicy(prv.Namespace, []string{"*"}),
					},
					commands: []cli.TestScenario{
						createTokenPolicyScenario(pub, "", "./tmp", "previous", true),
					},
				}, {
					name: "disallow-and-create-link",
					pubPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowIncomingLinkPolicy(pub.Namespace, false),
					},
					commands: []cli.TestScenario{
						createLinkTestScenario(prv, "", "previous"),
						linkStatusTestScenario(prv, "", "previous", false),
					},
					pubGetCheck: policyGetCheck{
						allowIncoming: &_false,
					},
				}, {
					name: "re-allow-and-check-link",
					pubPolicy: []skupperv1.SkupperClusterPolicySpec{
						allowIncomingLinkPolicy(pub.Namespace, true),
					},
					commands: []cli.TestScenario{
						linkStatusTestScenario(prv, "now", "previous", true),
						sitesConnectedTestScenario(pub, prv, "", "previous"),
						linkDeleteTestScenario(prv, "", "previous"),
					},
					pubGetCheck: policyGetCheck{
						allowIncoming: &_true,
					},
				},
			},
		}, {
			name: "cleanup",
			steps: []policyTestStep{
				{
					name:     "delete",
					parallel: true,
					commands: []cli.TestScenario{
						deleteSkupperTestScenario(pub, "pub"),
						deleteSkupperTestScenario(prv, "prv"),
					},
				},
			},
		},
	}

	policyTestRunner{
		scenarios: testTable,
	}.run(t, pub, prv)

}
