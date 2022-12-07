package tester

import (
	"fmt"
	"log"
	"regexp"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

type CliLinkStatus struct {
	execute.CliLinkStatus

	// List of links that should show on the status
	Outgoing       []CliLinkStatusOutgoing
	StrictOutgoing bool // instead of noOutgoing: list of links must be exact.

	// List of links that should show on the status
	Incoming       []CliLinkStatusIncoming
	StrictIncoming bool // instead of noOutgoing: list of links must be exact.

	// An optional retry profile.  If not given, a default will be used.  To
	// run just once, give it an empty frame2.RetryOptions{} struct
	*frame2.RetryOptions
}

func (cls CliLinkStatus) Execute() error {
	log.Printf("tester.CliLinkStatus")
	//log.Printf("%#v", cls)

	retryOptions := frame2.RetryOptions{
		Allow:  10,
		Ignore: 10,
	}
	if cls.RetryOptions != nil {
		retryOptions = *cls.RetryOptions
	}

	if cls.CliLinkStatus.Cmd.CmdResult == nil {
		// we're using this, so, if the client has not set it, we need to
		cls.CliLinkStatus.Cmd.CmdResult = &execute.CmdResult{}
	}

	retry := frame2.Retry{
		Options: retryOptions,
		Fn: func() error {

			err := cls.CliLinkStatus.Execute()
			if err != nil {
				return err
			}
			e := cli.Expect{}

			stdout, stderr := cls.CliLinkStatus.Cmd.CmdResult.Stdout, cls.CliLinkStatus.Cmd.CmdResult.Stderr
			// Outgoing
			if cls.StrictOutgoing && len(cls.Outgoing) == 0 {
				e.StdOut = append(e.StdOut, "There are no links configured or active")
				e.StdOutReNot = append(e.StdOutReNot, *regexp.MustCompile("^Link.*is.*active$"))

			}
			for _, i := range cls.Outgoing {
				var be string
				if i.Active {
					be = "is"
				} else {
					be = "not"
				}
				sentence := fmt.Sprintf("Link %v %v active", i.Name, be)
				re, err := regexp.Compile(sentence)
				if err != nil {
					return fmt.Errorf("tester.CliLinkStatus configuration error - the generated regexp %q is not valid", sentence)
				}
				e.StdOutRe = append(e.StdOutRe, *re)
			}

			// Incoming
			if cls.StrictIncoming && len(cls.Incoming) == 0 {
				e.StdOut = append(e.StdOut, "There are no active links")
				// TODO
				e.StdOutReNot = append(e.StdOutReNot, *regexp.MustCompile("^Link.*is.*active$"))

			}
			for _, i := range cls.Incoming {
				var be string
				if i.Active {
					be = "is"
				} else {
					be = "is not"
				}
				sentence := fmt.Sprintf("A link from the namespace %v .* %v active", i.SourceNamespace, be)
				re, err := regexp.Compile(sentence)
				if err != nil {
					return fmt.Errorf("tester.CliLinkStatus configuration error - the generated regexp %q is not valid", sentence)
				}
				e.StdOutRe = append(e.StdOutRe, *re)
			}

			// General
			e.StdErrReNot = append(e.StdErrReNot, *regexp.MustCompile(`\W`))
			return e.Check(stdout, stderr)
		},
	}

	_, err := retry.Run()
	return err
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
