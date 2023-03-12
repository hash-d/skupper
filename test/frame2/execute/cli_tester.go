package execute

import (
	"log"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

// This is a wrapper to the types available on test/utils/skupper/cli/
type CliTester struct {
	Tester  cli.SkupperCommandTester
	Cluster base.ClusterContext
}

func (c CliTester) Execute() error {
	log.Printf("CliTester: %+#v", c.Tester)
	stdout, stderr, err := c.Tester.Run(types.PlatformKubernetes, &c.Cluster)

	log.Printf("CliTester result: %v", err)
	log.Printf("CliTester:\nSTDOUT:\n%v\nSTDERR:\n%v", stdout, stderr)

	return err
}
