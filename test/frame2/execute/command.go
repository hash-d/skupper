package execute

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

const CmdDefaultTimeout = 2 * time.Minute

// Executes a command locally (ie, on the machine executing
// the test).
//
// If a CmdResult is provided by the caller, it will be populated
// with the stdout, stderr and error returned by the exec.Command
// execution, for furthe processing.
//
// Yet, most output check should be possible using the provided
// cli.Expect configuration.
//
// If both AcceptReturn and FailReturn are defined and the return
// status is not present on either, an error will be returned
type Cmd struct {
	// The command to be executed, as if exec.Command() had been called (ie, it
	// looks for the command on the PATH, if no slashes on it).  If empty, then
	// Cmd.Path must be set
	Command string
	// The exec.Cmd structure to be run.  If Command above is empty, it will be
	// used as is.  If Command is non-empty,  however, we'll run replace Cmd.Path
	// and Cmd.Args with those returned by exec.Command (ie, we'll let Go find
	// the path to the command).
	exec.Cmd
	Ctx     context.Context
	Timeout time.Duration // If not provided, a default timeout of 2 min is used
	//Shell        bool
	cli.Expect         // Configures checks on Stdout and Stderr
	AcceptReturn []int // consider these return status as a success.  Default only 0
	FailReturn   []int // Fail on any of these return status.  Default anything other than 0

	*CmdResult
}

type CmdResult struct {
	Stdout string
	Stderr string
	Err    error
}

type CmdError struct {
	Cmd    error
	Expect error
}

func (ce CmdError) Error() string {
	return fmt.Sprintf("cmd: %s, expect: %s", ce.Cmd, ce.Expect)
}

// Change this by Go 1.18's generic slices.Contains?
func containsInt(needle int, haystack []int) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false

}

func (c Cmd) Execute() error {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	ctx := c.Ctx
	// If no Context given, let's have a safe timeout
	if ctx == nil {
		ctx = context.Background()
	}

	// We'll set a context with timeout in two cases:
	// - For nil contexts
	// - For explicit requests
	// If nil context and no explicit timeout request, we set a default
	if c.Ctx == nil || c.Timeout > 0 {
		var timeout time.Duration

		if c.Timeout > 0 {
			timeout = c.Timeout
		} else {
			timeout = CmdDefaultTimeout
		}

		ctx_, fn := context.WithTimeout(context.Background(), timeout)
		ctx = ctx_
		defer fn()
	}

	// this will not be run; it's only used to prepare exec.Cmd.Path and exec.Cmd.Args
	var tmpcmd exec.Cmd
	// Preparing the command to run
	if c.Command == "" {
		// No command specified; we'll use the exec.Cmd structure as-is, just
		// overriding the context
		tmpcmd = c.Cmd
	} else {
		tmpcmd = *exec.Command(c.Command, c.Cmd.Args...)
	}
	log.Printf("original: %#v", c.Cmd)
	log.Printf("tmpcmp:   %#v", tmpcmd)

	// First, just give me exec.Cmd with a context, with an already-resolved path
	cmd := exec.CommandContext(ctx, tmpcmd.Path)
	// Now let's copy everything else from tmpcmd into it.
	mergo.Merge(&cmd, tmpcmd, mergo.WithOverride)
	// mergo will not merge Args, so we have to force it
	cmd.Args = tmpcmd.Args

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	log.Printf("Actual   : %#v", cmd)

	// Running the skupper command
	log.Printf("Running: %s %s\n", c.Command, strings.Join(c.Args, " "))
	cmdErr := cmd.Run()

	if c.CmdResult != nil {
		c.CmdResult.Stdout = stdout.String()
		c.CmdResult.Stderr = stdout.String()
		c.CmdResult.Err = cmdErr
	}

	if base.IsVerboseCommandOutput() {
		fmt.Printf("STDOUT:\n%v\n", stdout.String())
		fmt.Printf("STDERR:\n%v\n", stderr.String())
		fmt.Printf("Error: %v\n", cmdErr)
	}

	expectErr := c.Expect.Check(stdout.String(), stderr.String())

	var returnedCmdError error

	// The nil test is required, and must be outside of the type assertion.
	// Otherwise, go makes cmdErr assuma a nil ExitError form within the if,
	// and that does not count as a true nil
	if cmdErr != nil {
		exitError, ok := cmdErr.(*exec.ExitError)
		if ok {
			ret := exitError.ExitCode()
			if len(c.AcceptReturn) != 0 {
				if len(c.FailReturn) != 0 {
					// Both lists set
					switch {
					case containsInt(ret, c.AcceptReturn) && containsInt(ret, c.FailReturn):
						returnedCmdError = fmt.Errorf("cmd configuration error - the exit code %d is on both accept and fail lists: %w", ret, cmdErr)
					case containsInt(ret, c.AcceptReturn):
						returnedCmdError = nil
					case containsInt(ret, c.FailReturn):
						returnedCmdError = cmdErr
					default:
						returnedCmdError = fmt.Errorf("cmd configuration error - the exit code %d is on either accept nor fail lists: %w", ret, cmdErr)
					}
				} else {
					// Only AcceptReturn set
					if containsInt(ret, c.AcceptReturn) {
						returnedCmdError = nil
					} else {
						returnedCmdError = cmdErr
					}
				}
			} else {
				if len(c.FailReturn) != 0 {
					// Only FailReturn set
					if containsInt(ret, c.FailReturn) {
						returnedCmdError = cmdErr
					} else {
						returnedCmdError = nil
					}
				} else {
					// Neither list set; we can simply return what we got
					if cmdErr == nil {
						returnedCmdError = nil
					} else {
						returnedCmdError = cmdErr
					}
				}
			}
		} else {
			// Something happened outside of the realm of just getting the
			// command's exit code
			returnedCmdError = exitError
		}
	}

	var err error
	// It's only an error if either side is an error
	if returnedCmdError != nil || expectErr != nil {
		err = CmdError{
			Cmd:    returnedCmdError,
			Expect: expectErr,
		}
	}

	return err
}
