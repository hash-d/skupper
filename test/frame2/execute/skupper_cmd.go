package execute

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

// If both Namespace and ClusterContext are empty, the command will be executed
// without --namespace
type CliSkupper struct {
	Args []string

	// The primary way to define the namespace
	Namespace string

	// Secondary way to get the namespace, used only if Namespace is empty
	ClusterContext *base.ClusterContext

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd Cmd

	path string

	frame2.DefaultRunDealer
	frame2.Log
}

func (c *CliSkupper) Validate() error {
	return c.Execute()
}

func (cs *CliSkupper) Execute() error {
	log.Printf("execute.CliSkupper %v", cs.Args)
	//	log.Printf("%#v", cs)
	baseArgs := []string{}

	// TODO change this when adding Podman to frame2
	baseArgs = append(baseArgs, "--platform", "kubernetes")

	if cs.ClusterContext != nil {
		baseArgs = append(baseArgs, "--kubeconfig", cs.ClusterContext.KubeConfig)
	}

	if cs.Namespace != "" {
		baseArgs = append(baseArgs, "--namespace", cs.Namespace)
	} else {
		if cs.ClusterContext != nil {
			baseArgs = append(baseArgs, "--namespace", cs.ClusterContext.Namespace)
		}
	}
	cmd := cs.Cmd
	cmd.Command = cs.path
	if cmd.Command == "" {
		cmd.Command = "skupper"
	}
	cmd.Cmd.Args = append(baseArgs, cs.Args...)

	err := cmd.Execute()
	if err != nil {
		log.Printf("CmdResult: %#v", cmd.CmdResult)
		return fmt.Errorf("execute.CliSkupper: %w", err)
	}
	return nil
}

func (c *CliSkupper) SetSkupperCliPath(path string, env []string) {
	c.path = path
	c.Cmd.AdditionalEnv = env
}
