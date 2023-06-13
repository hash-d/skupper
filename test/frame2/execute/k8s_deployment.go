package execute

// try to keep this file in sync with ocp_deploymentconfig

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

	Ctx context.Context

	Result *appsv1.Deployment

	frame2.DefaultRunDealer
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
						Namespace: d.Namespace,
						Name:      d.Name,
					},
					ValidatorRetry: frame2.RetryOptions{
						// The pod can get started and die a few seconds later.
						// Here, we ensure it lived for a minimal time.
						// TODO make this configurable
						Ensure:     10,
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

	frame2.DefaultRunDealer
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
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	Result *appsv1.Deployment

	frame2.Log
	frame2.DefaultRunDealer
}

func (kdg *K8SDeploymentGet) Validate() error {
	ctx := frame2.ContextOrDefault(kdg.Ctx)

	var err error
	kdg.Result, err = kdg.Namespace.VanClient.KubeClient.AppsV1().Deployments(kdg.Namespace.Namespace).Get(ctx, kdg.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %q: %w", kdg.Name, err)
	}

	if kdg.Result.Status.ReadyReplicas < 1 {
		return fmt.Errorf("deployment %q has no ready replicas", kdg.Name)
	}

	return nil
}

// Wait for the named deployment to be available.  By default, it
// waits for up to two minutes, and ensures that the deployment reports
// as ready for at least 10s.
//
// That behavior can be changed using the RetryOptions field. On that
// field, the Ctx field cannot be set; if a different timeout is desired,
// set it on the Action's Ctx itself, and it will be used for the
// RetryOptions.
type K8SDeploymentWait struct {
	Name      string
	Namespace *base.ClusterContext
	Ctx       context.Context

	// On this field, do not set the context.  Use the K8SDeploymentWait.Ctx,
	// instead, it will be used for the underlying Retry
	RetryOptions frame2.RetryOptions
	frame2.DefaultRunDealer
	*frame2.Log
}

func (w K8SDeploymentWait) Validate() error {
	if w.RetryOptions.Ctx != nil {
		panic("RetryOptions.Ctx cannot be set for K8SDeploymentWait")
	}
	retry := w.RetryOptions
	if retry.IsEmpty() {
		ctx, cancel := context.WithTimeout(frame2.ContextOrDefault(w.Ctx), time.Minute*2)
		defer cancel()
		retry = frame2.RetryOptions{
			Ctx:        ctx,
			KeepTrying: true,
			Ensure:     10,
		}
	}
	phase := frame2.Phase{
		Runner: w.GetRunner(),
		Doc:    fmt.Sprintf("Waiting for deployment %q on ns %q", w.Name, w.Namespace.Namespace),
		MainSteps: []frame2.Step{
			{
				// TODO: stuff within functions need their runners replaced?
				ValidatorRetry: retry,
				Validator: &Function{
					Fn: func() error {
						validator := &K8SDeploymentGet{
							Namespace: w.Namespace,
							Name:      w.Name,
						}
						inner1 := frame2.Phase{
							Runner: w.GetRunner(),
							Doc:    fmt.Sprintf("Get the deployment %q on ns %q", w.Name, w.Namespace.Namespace),
							MainSteps: []frame2.Step{
								{
									Validator: validator,
								},
							},
						}
						err := inner1.Run()
						if err != nil {
							return err
						}

						inner2 := frame2.Phase{
							Runner: w.GetRunner(),
							Doc:    fmt.Sprintf("Check that the deployment %q is ready", w.Name),
							MainSteps: []frame2.Step{
								{
									Validator: &Function{
										Fn: func() error {
											if validator.Result == nil {
												return fmt.Errorf("deployment not ready: result is nil")
											}
											if validator.Result.Status.ReadyReplicas == 0 {
												return fmt.Errorf("deployment not ready: ready replicas is 0")
											}
											return nil
										},
									},
								},
							},
						}
						return inner2.Run()
					},
				},
			},
		},
	}

	return phase.Run()
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
	frame2.DefaultRunDealer
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
