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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This simply makes a request to k8s.NewDeployment
//
// See K8SDeployment for a more complete interface
type K8SDeploymentOpts struct {
	Name           string
	Namespace      *base.ClusterContext
	DeploymentOpts k8s.DeploymentOpts
	Wait           time.Duration // Waits for the deployment to be ready.  Otherwise, returns as soon as the create instruction has been issued.  If the wait lapses, return an error.
	Runner         *frame2.Run

	Ctx context.Context

	Result *appsv1.Deployment
}

func (d *K8SDeploymentOpts) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)
	deployment, err := k8s.NewDeployment(d.Name, d.Namespace.Namespace, d.DeploymentOpts)
	if err != nil {
		return err
	}

	d.Result = deployment

	d.Result, err = d.Namespace.VanClient.KubeClient.AppsV1().Deployments(d.Namespace.Namespace).Create(ctx, deployment, v1.CreateOptions{})
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
					Validator: &K8SDeploymentGet{
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
// # See K8SDeploymentOpts for a simpler interface
//
// For an example/template on creating a *v1.Deployment by hand, check
// test/utils/base/cluster_context.go (k8s.NewDeployment)
type K8SDeployment struct {
	Namespace  *base.ClusterContext
	Deployment *appsv1.Deployment

	Result *appsv1.Deployment
	Ctx    context.Context
}

func (d *K8SDeployment) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	var err error
	d.Result, err = d.Namespace.VanClient.KubeClient.AppsV1().Deployments(d.Namespace.Namespace).Create(ctx, d.Deployment, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to create deployment: %w", err)
	}

	return nil
}

type K8SDeploymentGet struct {
	Runner    *frame2.Run
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	Result *appsv1.Deployment

	frame2.Log
}

func (kdg K8SDeploymentGet) Validate() error {
	ctx := frame2.ContextOrDefault(kdg.Ctx)

	var err error
	kdg.Result, err = kdg.Namespace.VanClient.KubeClient.AppsV1().Deployments(kdg.Namespace.Namespace).Get(ctx, kdg.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get deployment %q: %w", kdg.Name, err)
	}

	if kdg.Result.Status.ReadyReplicas < 1 {
		return fmt.Errorf("Deployment %q has no ready replicas", kdg.Name)
	}

	return nil
}

type K8SDeploymentAnnotate struct {
	Namespace   *base.ClusterContext
	Name        string
	Annotations map[string]string

	Ctx context.Context
}

func (kda K8SDeploymentAnnotate) Execute() error {
	ctx := frame2.ContextOrDefault(kda.Ctx)
	// Retrieving Deployment
	deploy, err := kda.Namespace.VanClient.KubeClient.AppsV1().Deployments(kda.Namespace.VanClient.Namespace).Get(ctx, kda.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}

	for k, v := range kda.Annotations {
		deploy.Annotations[k] = v
	}
	_, err = kda.Namespace.VanClient.KubeClient.AppsV1().Deployments(kda.Namespace.Namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	return err

}

type K8SUndeploy struct {
	Name      string
	Namespace *base.ClusterContext
	Wait      time.Duration // Waits for the deployment to be gone.  Otherwise, returns as soon as the delete instruction has been issued.  If the wait lapses, return an error.

	Ctx context.Context
}

func (k *K8SUndeploy) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)
	err := k.Namespace.VanClient.KubeClient.AppsV1().Deployments(k.Namespace.VanClient.Namespace).Delete(ctx, k.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	if k.Wait == 0 {
		return nil
	}
	retry := frame2.Retry{
		Options: frame2.RetryOptions{
			Ctx:        ctx,
			Timeout:    k.Wait,
			KeepTrying: true,
		},
		Fn: func() error {
			_, err := k.Namespace.VanClient.KubeClient.AppsV1().Deployments(k.Namespace.VanClient.Namespace).Get(ctx, k.Name, metav1.GetOptions{})
			if err == nil {
				return fmt.Errorf("deployment %v still available after deletion", k.Name)
			}
			return nil
		},
	}
	_, err = retry.Run()
	if err != nil {
		return err
	}
	return nil
}
