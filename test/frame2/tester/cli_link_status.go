package tester

import (
	"fmt"
	"regexp"

	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

type CliLinkStatus struct {
	execute.CliLinkStatus

	// List of links that should show on the status
	outgoing       []CliLinkStatusOutgoing
	StrictOutgoing bool // instead of noOutgoing: list of links must be exact.

	// List of links that should show on the status
	incoming       []CliLinkStatusIncoming
	StrictIncoming bool // instead of noOutgoing: list of links must be exact.
}

func (cls CliLinkStatus) Exec() error {
	err := cls.CliLinkStatus.Execute()
	if err != nil {
		return err
	}
	e := cli.Expect{}

	stdout, stderr := cls.CliLinkStatus.Cmd.CmdResult.Stdout, cls.CliLinkStatus.Cmd.CmdResult.Stderr
	if cls.StrictOutgoing && len(cls.outgoing) == 0 {
		e.StdOut = append(e.StdOut, "There are no links configured or active")
		e.StdOutReNot = append(e.StdOutReNot, *regexp.MustCompile("^Link.*is.*active$"))

	}
	for _, i := range cls.outgoing {
		var be string
		if i.Active {
			be = "is"
		} else {
			be = "is not"
		}
		sentence := fmt.Sprintf("^Link %v %v active$", i.Name, be)
		re, err := regexp.Compile(sentence)
		if err != nil {
			return fmt.Errorf("tester.CliLinkStatus configuration error - the generated regexp %q is not valid", sentence)
		}
		e.StdOutRe = append(e.StdOutRe, *re)
	}
	e.StdErrReNot = append(e.StdErrReNot, *regexp.MustCompile(`\W`))
	return e.Check(stdout, stderr)
}

// This is not a tester; it's just a configuration struct
type CliLinkStatusOutgoing struct {
	// Name will be interpreted as a regexp
	Name   string
	Active bool
}

// This is not a tester; it's just a configuration struct
type CliLinkStatusIncoming struct {
	SourceNamespace string
	Active          bool
}
