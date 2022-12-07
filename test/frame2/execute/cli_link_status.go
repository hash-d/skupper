package execute

import (
	"strconv"
	"time"
)

type CliLinkStatus struct {

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd

	// These will be put on the command right after the subcommands, and
	// before other options
	AdditionalArgs []string

	// Instead of keeping this documentation up to date, run
	//
	//   skupper link status -h
	//
	// for a description of each below.

	//	Failure ClaimFailure
	//	Name, Active

	// Wait and Timeotu are only set if non-zero.  If you want to test with
	// a zero value or something else, use AdditionalArgs
	Wait    int
	Timeout time.Duration
	Verbose bool
}

func (cls CliLinkStatus) Execute() error {
	var args = []string{"link", "status"}
	args = append(args, cls.AdditionalArgs...)

	if cls.Wait > 0 {
		args = append(args, "--wait", strconv.Itoa(cls.Wait))
	}

	if cls.Verbose {
		args = append(args, "--verbose")
	}

	if cls.Timeout > 0 {
		args = append(args, "--timeout", cls.Timeout.String())
	}

	cliSkupper := CliSkupper{
		Args: args,
		Cmd:  cls.Cmd,
	}
	return cliSkupper.Execute()
}
