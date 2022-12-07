package execute

import (
	"log"
	"strconv"
	"time"
)

type CliLinkStatus struct {

	// skupper link status options
	//
	// Instead of keeping this documentation up to date, run
	//
	//   skupper link status -h
	//
	// for a description of each below.
	//
	// Wait and Timeotu are only set if non-zero.  If you want to test with
	// a zero value or something else, use AdditionalArgs

	Wait    int           // skupper link status --help
	Timeout time.Duration // skupper link status --help
	Verbose bool          // skupper link status --help

	// These will be put on the command right after the subcommands, and
	// before other options selected above
	AdditionalArgs []string

	//	Failure ClaimFailure
	//	Name, Active

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	CliSkupper
}

func (cls CliLinkStatus) Execute() error {
	log.Printf("execute.CliLinkStatus")
	//log.Printf("%#v", cls)
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

	cliSkupper := cls.CliSkupper
	cliSkupper.Args = args
	return cliSkupper.Execute()
}
