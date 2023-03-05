package validates

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

// TODO - This should move to a validate package, but it can't be there
// right now because of an import cycle.  I don't want to refactor the
// code before I finish the est
//
// Executes nslookup within a pod, to check whether a name is valid
// within a namespace or cluster
type Nslookup struct {
	Namespace *base.ClusterContextPromise

	Name string

	Cmd execute.Cmd

	Runner *frame2.Run
}

func (n Nslookup) Validate() error {
	cc, err := n.Namespace.Satisfy()
	if err != nil {
		return err
	}

	arg := fmt.Sprintf("kubectl --namespace %s exec deploy/dnsutils -- nslookup %q", cc.Namespace, n.Name)

	n.Cmd.Command = arg
	n.Cmd.Shell = true

	phase := frame2.Phase{
		Runner: n.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &n.Cmd,
			},
		},
	}
	return phase.Run()
}