package execute

import (
	"os/exec"
)

type CliSkupper struct {
	Args []string

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd
}

func (cs *CliSkupper) Execute() error {
	cmd := Cmd{
		Command: "skupper",
		Cmd: exec.Cmd{
			Args: cs.Args,
		},
		Ctx:       cs.Ctx,
		Timeout:   cs.Timeout,
		CmdResult: cs.CmdResult,
	}
	cmd.Run()
	return nil
}
