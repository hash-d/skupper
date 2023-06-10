package execute

import (
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type Kubectl struct {
	Args []string

	// Secondary way to get the namespace, used only if Namespace is empty
	ClusterContext *base.ClusterContext

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd Cmd

	Runner frame2.Run
}

func (k Kubectl) Execute() error {

	// TODO: add --kubeconfig based on k.ClusterContext

	if k.Cmd.Shell {
		k.Cmd.Command = "kubectl " + k.Cmd.Command
	} else {
		k.Cmd.Command = "kubectl"
	}

	phase := frame2.Phase{
		MainSteps: []frame2.Step{
			{
				Modify: &k.Cmd,
			},
		},
	}

	return phase.Run()
}
