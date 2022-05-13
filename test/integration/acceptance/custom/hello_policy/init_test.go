//go:build policy
// +build policy

package hello_policy

import (
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

// Returns a test scenario that initializes skupper on the given context as interior,
// using the same arguments as used on the main hello_world test, and then confirms
// it is up and has policy enabled.
func skupperInitInteriorTestScenario(ctx *base.ClusterContext, prefix string, withPolicy bool) (initSteps cli.TestScenario) {
	initSteps = cli.TestScenario{
		Name: prefixName(prefix, "init-skupper-interior"),
		Tasks: []cli.SkupperTask{
			{Ctx: ctx, Commands: []cli.SkupperCommandTester{
				// skupper init - interior mode, enabling console and internal authentication
				&cli.InitTester{
					ConsoleAuth:         "internal",
					ConsoleUser:         "internal",
					ConsolePassword:     "internal",
					RouterMode:          "interior",
					EnableConsole:       true,
					EnableRouterConsole: true,
				},
				// skupper status - verify initialized as interior
				&cli.StatusTester{
					RouterMode:          "interior",
					ConsoleEnabled:      true,
					ConsoleAuthInternal: true,
					PolicyEnabled:       cli.Boolp(withPolicy),
				},
			}},
		},
	}
	return
}

// Returns a test scenario that initializes skupper on the given context as edge,
// using the same arguments as used on the main hello_world test, and then confirms
// it is up and has policy enabled.
func skupperInitEdgeTestScenario(ctx *base.ClusterContext, prefix string, withPolicy bool) (initSteps cli.TestScenario) {
	initSteps = cli.TestScenario{
		Name: prefixName(prefix, "init-skupper-edge"),
		Tasks: []cli.SkupperTask{
			{Ctx: ctx, Commands: []cli.SkupperCommandTester{
				// skupper init - edge mode, no console and unsecured
				&cli.InitTester{
					ConsoleAuth:         "unsecured",
					ConsoleUser:         "admin",
					ConsolePassword:     "admin",
					Ingress:             "none",
					RouterDebugMode:     "gdb",
					RouterLogging:       "trace",
					RouterMode:          "edge",
					SiteName:            "private",
					EnableConsole:       false,
					EnableRouterConsole: false,
					// ConsoleIngress:      "none",
				},
				// skupper status - verify initialized as edge
				&cli.StatusTester{
					RouterMode:    "edge",
					SiteName:      "private",
					PolicyEnabled: cli.Boolp(withPolicy),
				},
			}},
		},
	}
	return
}
