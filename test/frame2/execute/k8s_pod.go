package execute

import (
	"context"
	"fmt"
	"strings"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This will return a single pod.  If name is provided, that will be used.
// Otherwise, the pods with the given label will be listed, and the first
// in the list will be returned
type K8SPodGet struct {
	Namespace *base.ClusterContext
	Name      string
	Labels    map[string]string
	Ctx       context.Context

	Result *corev1.Pod

	frame2.Log
	frame2.DefaultRunDealer
}

func (g *K8SPodGet) Execute() error {
	ctx := frame2.ContextOrDefault(g.Ctx)

	if g.Name != "" {
		var err error
		g.Result, err = g.Namespace.VanClient.KubeClient.CoreV1().Pods(g.Namespace.Namespace).Get(ctx, g.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod %q by name: %w", g.Name, err)
		}
		return nil
	}

	var items []string
	// TODO: is there an API that already does that?
	for k, v := range g.Labels {
		items = append(items, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(items, ",")
	podList, err := g.Namespace.VanClient.KubeClient.CoreV1().Pods(g.Namespace.Namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: selector,
			Limit:         1,
		})
	if err != nil {
		return fmt.Errorf("failed to get pod list by labels: %w", err)
	}
	if len(podList.Items) != 1 {
		return fmt.Errorf("failed to get pod by labels")
	}
	g.Result = &podList.Items[0]

	return nil
}

type K8SPodExecute struct {
	Pod       *K8SPodGet
	Container string
	Command   []string
	Ctx       context.Context
	Expect    cli.Expect // Configures checks on Stdout and Stderr

	// TODO: use common code with execute.Command
	ForceOutput bool // Shows this command's output on log, regardless of environment config
	// ForceNoOutput bool       // No output, regardless of environment config.  Takes precedence over the above

	// These are probably not implementable; k8s.io/client-go/tools/remotecommand does not
	// return exit status
	//
	// AcceptReturn  []int      // consider these return status as a success.  Default only 0
	// FailReturn    []int      // Fail on any of these return status.  Default anything other than 0

	frame2.Log

	*CmdResult
}

func (e *K8SPodExecute) Execute() error {
	err := e.Pod.Execute()
	if err != nil {
		return fmt.Errorf("K8SPodExecute failed to get pod: %w", err)
	}

	stdout, stderr, err := k8s.Execute(
		e.Pod.Namespace.VanClient.KubeClient,
		e.Pod.Namespace.VanClient.RestConfig,
		e.Pod.Namespace.Namespace,
		e.Pod.Result.GetName(),
		e.Container,
		e.Command)
	e.CmdResult = &CmdResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if e.ForceOutput || base.IsVerboseCommandOutput() {
		e.Log.Printf("STDOUT:\n%v\n", stdout.String())
		e.Log.Printf("STDERR:\n%v\n", stderr.String())
		e.Log.Printf("Error: %v\n", err)
	}
	if err != nil {
		return fmt.Errorf("K8SPodExecute failed execution:  %w", err)
	}

	expectErr := e.Expect.Check(e.CmdResult.Stdout, e.CmdResult.Stderr)
	if expectErr != nil {
		return expectErr
	}

	return nil
}

func (e *K8SPodExecute) Validate() error {
	return e.Execute()
}
