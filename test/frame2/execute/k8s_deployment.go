package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This simply makes a request to k8s.NewDeployment
//
// See K8SDeployment for a more complete interface
type K8SDeploymentOpts struct {
	Name           string
	Namespace      *base.ClusterContextPromise
	DeploymentOpts k8s.DeploymentOpts
	Wait           time.Duration // Waits for the deployment to be ready.  Otherwise, returns as soon as the create instruction has been issued.  If the wait lapses, return an error.
	Runner         *frame2.Run

	Ctx context.Context

	Result *appsv1.Deployment
}

func (d *K8SDeploymentOpts) Execute() error {
	cc, err := d.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}
	deployment, err := k8s.NewDeployment(d.Name, cc.Namespace, d.DeploymentOpts)
	if err != nil {
		return err
	}

	d.Result = deployment

	d.Result, err = cc.VanClient.KubeClient.AppsV1().Deployments(cc.Namespace).Create(deployment)
	if err != nil {
		return fmt.Errorf("Failed to create deployment: %w", err)
	}

	if d.Wait > 0 {
		ctx := d.Ctx
		var fn context.CancelFunc
		if ctx == nil {
			ctx, fn = context.WithTimeout(context.Background(), d.Wait)
			defer fn()
		}

		phase := frame2.Phase{
			Runner: d.Runner,
			MainSteps: []frame2.Step{
				{
					Validator: K8SDeploymentGet{
						Runner:    d.Runner,
						Namespace: d.Namespace,
						Name:      d.Name,
					},
					ValidatorRetry: frame2.RetryOptions{
						Ctx:        ctx,
						KeepTrying: true,
					},
				},
			},
		}
		return phase.Run()
	}

	return nil
}

// Executes a fully specified K8S deployment
//
// See K8SDeploymentOpts for a simpler interface
//
// For an example/template on creating a *v1.Deployment by hand, check
// test/utils/base/cluster_context.go (k8s.NewDeployment)
//
type K8SDeployment struct {
	Namespace  *base.ClusterContextPromise
	Deployment *appsv1.Deployment

	Result *appsv1.Deployment
}

func (d *K8SDeployment) Execute() error {
	cc, err := d.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}

	d.Result, err = cc.VanClient.KubeClient.AppsV1().Deployments(cc.Namespace).Create(d.Deployment)
	if err != nil {
		return fmt.Errorf("Failed to create deployment: %w", err)
	}

	return nil
}

type K8SDeploymentGet struct {
	Runner    *frame2.Run
	Namespace *base.ClusterContextPromise
	Name      string

	Result *appsv1.Deployment
}

func (kdg K8SDeploymentGet) Validate() error {
	cc, err := kdg.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}

	kdg.Result, err = cc.VanClient.KubeClient.AppsV1().Deployments(cc.Namespace).Get(kdg.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get deployment %q: %w", kdg.Name, err)
	}

	if kdg.Result.Status.ReadyReplicas < 1 {
		return fmt.Errorf("Deployment %q has no ready replicas", kdg.Name)
	}

	return nil
}
