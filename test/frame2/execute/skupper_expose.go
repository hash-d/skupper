package execute

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

// TODO: rename this to CLI; make a general type that can call
// the CLI, create annotations, use Ansible or site controller,
// per configuration.
type SkupperExpose struct {
	Namespace *base.ClusterContext
	Type      string
	Name      string

	// TODO.  Change this into some constants, so it can be reused and translated by different backing
	//        implementations.
	// A string that will compile into a Regex, which matches the command stderr to define an
	// expected failure.
	FailureReason string

	Address                string
	Headless               bool
	Protocol               string
	Ports                  []int
	PublishNotReadyAddress bool
	TargetPorts            []string

	AutoTeardown bool

	frame2.DefaultRunDealer
}

func (se SkupperExpose) Execute() error {

	var args []string

	if se.Type == "" || se.Name == "" {
		return fmt.Errorf("SkupperExpose configuration error - type and name must be specified")
	}

	args = append(args, "expose", se.Type, se.Name)

	if se.Headless {
		args = append(args, "--headless")
	}

	if se.PublishNotReadyAddress {
		args = append(args, "--publish-not-ready-addresses")
	}

	if se.Address != "" {
		args = append(args, "--address", se.Address)
	}

	if se.Protocol != "" {
		args = append(args, "--protocol", se.Protocol)
	}

	if len(se.TargetPorts) != 0 {
		args = append(args, "--target-port", strings.Join(se.TargetPorts, ","))
	}

	if len(se.Ports) != 0 {
		var tmpPorts []string
		for _, p := range se.Ports {
			tmpPorts = append(tmpPorts, strconv.Itoa(p))
		}
		args = append(args, "--port", strings.Join(tmpPorts, ","))
	}

	cmd := Cmd{}

	if se.FailureReason != "" {
		cmd.FailReturn = []int{0}
		re, err := regexp.Compile(se.FailureReason)
		if err != nil {
			return fmt.Errorf("SkupperExpose failed to compile FailureReason %q as a regexp: %w", se.FailureReason, err)
		}
		cmd.Expect = cli.Expect{
			StdErrRe: []regexp.Regexp{*re},
		}
	}

	phase := frame2.Phase{
		Runner: se.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: se.Namespace,
					Cmd:            cmd,
				},
			},
		},
	}

	return phase.Run()
}

func (se SkupperExpose) Teardown() frame2.Executor {

	if !se.AutoTeardown {
		return nil
	}

	return SkupperUnexpose{
		Namespace: se.Namespace,
		Type:      se.Type,
		Name:      se.Name,
		Address:   se.Address,

		Runner: se.Runner,
	}

}

type SkupperUnexpose struct {
	Namespace *base.ClusterContext
	Type      string
	Name      string
	Address   string

	Runner *frame2.Run
}

func (su SkupperUnexpose) Execute() error {
	var args []string

	if su.Type == "" || su.Name == "" {
		return fmt.Errorf("SkupperExpose configuration error - type and name must be specified")
	}

	args = append(args, "unexpose", su.Type, su.Name)

	if su.Address != "" {
		args = append(args, "--address", su.Address)
	}

	phase := frame2.Phase{
		Runner: su.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: su.Namespace,
				},
			},
		},
	}

	return phase.Run()

}
