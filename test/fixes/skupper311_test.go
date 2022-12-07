//go:build fixes
// +build fixes

package fixes

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/walk"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli/link"
)

func Test311(t *testing.T) {
	var runner = &base.ClusterTestRunnerBase{}
	var pub = runner.GetPublicContextPromise(1)
	var prv = runner.GetPrivateContextPromise(1)

	var tests = frame2.TestRun{
		Name: "Test311",
		Setup: []frame2.Step{
			{
				Modify: walk.SegmentSetup{
					Namespace: pub,
				},
			},
		},
		Teardown: []frame2.Step{
			{
				Modify: walk.SegmentTeardown{},
			},
		},
		MainSteps: []frame2.Step{
			{
				Name: "setup-verify",
				Substeps: []*frame2.Step{
					{
						Modify: execute.CliTester{
							Cluster: *prv,
							Tester: &link.StatusTester{
								Name:    "public",
								Active:  true,
								Timeout: 10 * time.Second,
							},
						},
					}, {
						Modify: execute.CliTester{
							Cluster: *pub,
							Tester: &link.StatusTester{
								Timeout: 10 * time.Second,
							},
						},
					},
				},
			},
		},
	}

	tests.Run(t)

}
