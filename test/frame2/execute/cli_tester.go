package execute

import (
	"log"

	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

type CliTester struct {
	Tester  cli.SkupperCommandTester
	Cluster base.ClusterContextPromise
}

func (c CliTester) Execute() error {
	log.Printf("CliTester: %+#v", c.Tester)
	cluster, err := c.Cluster.Satisfy()
	if err != nil {
		return err
	}
	stdout, stderr, err := c.Tester.Run(cluster)

	log.Printf("CliTester result: %v", err)
	log.Printf("CliTester:\nSTDOUT:\n%v\nSTDERR:\n%v", stdout, stderr)

	return err
}
