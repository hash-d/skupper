//go:build meta_test
// +build meta_test

package execute

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

var unknownCommand = "you-dont-have-a-command-with-this-name-do-you"
var doneCtx, _ = context.WithTimeout(context.Background(), time.Microsecond)

var resultCommunication = CmdResult{}

func TestCmd(t *testing.T) {
	tests.RunT(t)
}

type cmdValidator struct {
	cmd                 Cmd
	errorOnCommand      bool
	errorOnExpect       bool
	nonCmdErr           bool
	resultCommunication *CmdResult

	frame2.Log
}

func (ct cmdValidator) Validate() error {
	err := ct.cmd.Execute()

	log.Printf("CmdValidator checking the error %v", err)

	foundErrors := []string{}
	if err == nil {
		// No error and I was not expecting anything anyway
		if !ct.errorOnCommand && !ct.errorOnExpect {
			return nil
		}
		if ct.errorOnCommand {
			foundErrors = append(foundErrors, "expected command to fail, but it didn't")
		}
		if ct.errorOnExpect {
			foundErrors = append(foundErrors, "expected Expect to fail, but it didn't")
		}
	} else {
		typedErr, ok := err.(CmdError)
		if !ok {
			// This is not a CmdError
			if !ct.nonCmdErr {
				// And I was expecting a CmdError
				foundErrors = append(foundErrors, fmt.Sprintf("returned error was not of type CmdError (%T)", err))
			}
		} else {
			// This is a CmdError
			if ct.nonCmdErr {
				// And I was expecting something else.
				foundErrors = append(foundErrors, fmt.Sprintf("unexpected error expected, but got %v", typedErr))
			}
			if ct.errorOnExpect {
				if typedErr.Expect == nil {
					foundErrors = append(foundErrors, "expected Expect to fail, but it didn't")
				}
			} else {
				if typedErr.Expect != nil {
					foundErrors = append(foundErrors, fmt.Sprintf("unexpected Expect failure: %v", typedErr))
				}
			}
			if ct.errorOnCommand {
				if typedErr.Cmd == nil {
					foundErrors = append(foundErrors, "expected command to fail, but it didn't")
				}
			} else {
				if typedErr.Cmd != nil {
					foundErrors = append(foundErrors, fmt.Sprintf("unexpected command failure: %v", typedErr))
				}
			}
		}

	}

	if len(foundErrors) > 0 {
		//		log.Printf("Command failed: %+v", ct.cmd)
		return fmt.Errorf(
			"cmdValidator failed: %s",
			strings.Join(foundErrors, "; "),
		)
	}

	return nil
}

var tests = frame2.Phase{
	Name: "test-command",
	MainSteps: []frame2.Step{
		{
			Name: "positive-expected",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "true",
				},
			},
		}, {
			Name: "negative-unexpected",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "false",
				},
			},
			ExpectError: true,
		}, {
			Name: "positive-unexpected",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "true",
				},
				errorOnCommand: true,
			},
			ExpectError: true,
		}, {
			Name: "negative-expected",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:    "false",
					FailReturn: []int{0},
				},
			},
		}, {
			Name: "args-look",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "echo",
					Cmd: exec.Cmd{
						Args: []string{"Hello"},
					},
					Expect: cli.Expect{
						StdOut: []string{"Hello"},
					},
				},
			},
		}, {
			Name: "args-look-not-found",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: unknownCommand,
				},
				errorOnCommand: true,
			},
		}, {
			Name: "args-specific",
			Doc:  "Command, not Path, but with slashes",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "/usr/bin/echo",
					Cmd: exec.Cmd{
						Args: []string{"Hello"},
					},
					Expect: cli.Expect{
						StdOut: []string{"Hello"},
					},
				},
			},
		}, {
			Name: "args-specific-not-found",
			Doc:  "Command, not Path, but with slashes; not found",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "/usr/bin/" + unknownCommand,
					Cmd: exec.Cmd{
						Args: []string{"Hello"},
					},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "path-args",
			Validator: &cmdValidator{
				cmd: Cmd{
					Cmd: exec.Cmd{
						Path: "/usr/bin/echo",
						// If you're using path, you have to ensure yourself that
						// you're using arg[0] as the command name
						Args: []string{"myecho", "Hello"},
					},
					Expect: cli.Expect{
						StdOut:      []string{"Hello"},
						StdOutReNot: []regexp.Regexp{*regexp.MustCompile("myecho")},
					},
				},
			},
		}, {
			Name: "path-no-args",
			Validator: &cmdValidator{
				cmd: Cmd{
					Cmd: exec.Cmd{
						Path: "/usr/bin/sleep",
						// If you're using path, you have to ensure yourself that
						// you're using arg[0] as the command name
						Args: []string{"mysleep"},
					},
					Expect: cli.Expect{
						// We have no arguments, but we set the name as "mysleep",
						// so it should show on the error
						StdErr: []string{"mysleep"},
					},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "empty-shell",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "",
					Shell:   true,
				},
				// This should be the only situation where an error other than
				// CmdErr is returned
				nonCmdErr: true,
			},
		}, {
			Name: "shell-args",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "echo hello",
					Shell:   true,
					Expect: cli.Expect{
						StdOut: []string{"hello"},
					},
				},
			},
		}, {
			Name: "shell-no-args",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "cal",
					Shell:   true,
					Expect: cli.Expect{
						// This might fail sporadically after 2100...
						StdOut: []string{" 20"},
					},
				},
			},
		}, {
			Name: "shell-not-found",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: unknownCommand,
					Shell:   true,
					Expect: cli.Expect{
						StdErr: []string{unknownCommand},
					},
					AcceptReturn: []int{127}, // 127 is for command not found
				},
			},
		}, {
			Name: "non-exit-error-failure",
			Doc:  "With non-ExitError failure, do we get strings, or does stuff fail?",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "date",
					Expect: cli.Expect{
						// Does Expect fail in some way, because the output
						// is not available?
						StdOut: []string{""},
					},
					// This falses the command to fail, with a non-ExitError return
					Ctx: doneCtx,
				},
				errorOnExpect:  false,
				errorOnCommand: true,
			},
		}, {
			Name: "timeout",
			Doc:  "Even with a timeout, we should get the output",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "echo hello; /usr/bin/sleep 60",
					Shell:   true,
					Timeout: time.Second,
					Expect: cli.Expect{
						StdOut: []string{"hello"},
					},
				},
				errorOnExpect:  false,
				errorOnCommand: true,
			},
		}, {
			Name: "accept-list-ok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 2",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3},
				},
			},
		}, {
			Name: "accept-list-nok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 4",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "accept-list-nok-with-zero",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 0",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3},
				},
			},
		}, {
			Name: "fail-list-ok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:    "exit 4",
					Shell:      true,
					FailReturn: []int{1, 2, 3},
				},
			},
		}, {
			Name: "fail-list-nok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:    "exit 2",
					Shell:      true,
					FailReturn: []int{1, 2, 3},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "both-lists-fail-ok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 4",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3},
					FailReturn:   []int{4, 5, 6},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "both-lists-succes-ok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 2",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3},
					FailReturn:   []int{4, 5, 6},
				},
			},
		}, {
			Name: "both-lists-neither-nok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 100",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3},
					FailReturn:   []int{4, 5, 6},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "both-lists-both-nok",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command:      "exit 3",
					Shell:        true,
					AcceptReturn: []int{1, 2, 3, 4},
					FailReturn:   []int{3, 4, 5, 6},
				},
				errorOnCommand: true,
			},
		}, {
			Name: "TODO",
			Doc:  "resultCommunication and stdout/stderr supplied by the user",
			Validator: &cmdValidator{
				cmd: Cmd{
					Command: "false",
				},
			},
		},
	},
}
