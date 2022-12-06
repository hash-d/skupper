//go:build meta_test
// +build meta_test

package execute

import (
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

func TestCmd(t *testing.T) {
	tests.Run(t)
}

type cmdValidator struct {
	cmd          Cmd
	commandError bool
	expectError  bool
}

func (ct cmdValidator) Validate() error {
	err := ct.cmd.Execute()

	log.Printf("Checking the error %v", err)

	foundErrors := []string{}
	if err == nil {
		// No error and I was not expecting anything anyway
		if !ct.commandError && !ct.expectError {
			return nil
		}
		if ct.commandError {
			foundErrors = append(foundErrors, "expected command to fail, but it didn't")
		}
		if ct.expectError {
			foundErrors = append(foundErrors, "expected Expect to fail, but it didn't")
		}
	} else {
		typedErr, ok := err.(CmdError)
		if !ok {
			foundErrors = append(foundErrors, fmt.Sprintf("returned error was not of type CmdError (%T)", err))
		} else {
			if ct.expectError {
				if typedErr.Expect == nil {
					foundErrors = append(foundErrors, "expected Expect to fail, but it didn't")
				}
			} else {
				if typedErr.Expect != nil {
					foundErrors = append(foundErrors, fmt.Sprintf("unexpected Expect failure: %v", ct.expectError))
				}
			}
			if ct.commandError {
				if typedErr.Cmd == nil {
					foundErrors = append(foundErrors, "expected command to fail, but it didn't")
				}
			} else {
				if typedErr.Cmd != nil {
					foundErrors = append(foundErrors, fmt.Sprintf("unexpected command failure: %v", ct.commandError))
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

var tests = frame2.TestRun{
	Name: "test-command",
	MainSteps: []frame2.Step{
		{
			Name: "positive-expected",
			Validator: cmdValidator{
				cmd: Cmd{
					Command: "true",
				},
			},
		}, {
			Name: "negative-unexpected",
			Validator: cmdValidator{
				cmd: Cmd{
					Command: "false",
				},
			},
			ExpectError: true,
		}, {
			Name: "positive-unexpected",
			Validator: cmdValidator{
				cmd: Cmd{
					Command: "true",
				},
				commandError: true,
			},
			ExpectError: true,
		}, {
			Name: "negative-expected",
			Validator: cmdValidator{
				cmd: Cmd{
					Command:    "false",
					FailReturn: []int{0},
				},
			},
		}, {
			Name: "args-look",
			Validator: cmdValidator{
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
			Name: "args-specific",
			Validator: cmdValidator{
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
			Name: "path-args",
			Validator: cmdValidator{
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
			Validator: cmdValidator{
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
				commandError: true,
			},
		}, {
			Name: "context",
			Validator: cmdValidator{
				cmd: Cmd{
					Command: "/usr/bin/sleep",
					Cmd: exec.Cmd{
						Args: []string{"60"},
					},
					Timeout: time.Millisecond,
					Expect: cli.Expect{
						StdErrReNot: []regexp.Regexp{*regexp.MustCompile(`\W+`)},
					},
				},
				expectError:  false,
				commandError: true,
			},
		},
	},
}
