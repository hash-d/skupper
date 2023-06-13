package execute

import (
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type OcCli struct {
	Args []string

	ClusterContext *base.ClusterContext

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd Cmd

	frame2.DefaultRunDealer
}

func (k OcCli) Execute() error {

	// TODO: add --kubeconfig based on k.ClusterContext

	if k.Cmd.Shell {
		k.Cmd.Command = "oc " + k.Cmd.Command
	} else {
		k.Cmd.Command = "oc"
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
