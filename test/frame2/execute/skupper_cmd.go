package execute

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/utils/base"
)

// If both Namespace and ClusterContext are empty, the command will be executed
// without --namespace
type CliSkupper struct {
	Args []string

	// The primary way to define the namespace
	Namespace string

	// Secondary way to get the namespace, used only if Namespace is empty
	ClusterContext *base.ClusterContextPromise

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd
}

func (cs *CliSkupper) Execute() error {
	log.Printf("execute.CliSkupper")
	//	log.Printf("%#v", cs)
	baseArgs := []string{}
	if cs.Namespace != "" {
		baseArgs = append(baseArgs, "--namespace", cs.Namespace)
	} else {
		if cs.ClusterContext != nil {
			namespace, err := cs.ClusterContext.Satisfy()
			if err != nil {
				return fmt.Errorf("CliSkupper failed getting the namespace: %w", err)
			}
			baseArgs = append(baseArgs, "--namespace", namespace.Namespace)
		}
	}
	cmd := cs.Cmd
	cmd.Command = "skupper"
	cmd.Cmd.Args = append(baseArgs, cs.Args...)

	err := cmd.Execute()
	if err != nil {
		log.Printf("CmdResult: %#v", cmd.CmdResult)
		return fmt.Errorf("execute.CliSkupper: %w", err)
	}
	return nil
}
