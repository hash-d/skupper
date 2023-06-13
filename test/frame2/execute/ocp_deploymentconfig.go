package execute

// Try to keep this file in sync with k8s_deployment

import (
	"context"
	"fmt"
	"time"

	osappsv1 "github.com/openshift/api/apps/v1"
	clientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// See OCPDeploymentConfig for a more complete interface
type OCPDeploymentConfigOpts struct {
	Name           string
	Namespace      *base.ClusterContext
	DeploymentOpts k8s.DeploymentOpts
	Wait           time.Duration // Waits for the deployment to be ready.  Otherwise, returns as soon as the create instruction has been issued.  If the wait lapses, return an error.

	Ctx context.Context

	Result *osappsv1.DeploymentConfig

	frame2.DefaultRunDealer
}

//          Image         string
//          Labels        map[string]string
//          RestartPolicy v12.RestartPolicy
//          Command       []string
//          Args          []string
//          EnvVars       []v12.EnvVar
//          ResourceReq   v12.ResourceRequirements
//          SecretMounts  []SecretMount

// TODO: remove this whole thing?
func (d *OCPDeploymentConfigOpts) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	var volumeMounts []v1.VolumeMount
	var volumes []v1.Volume

	for _, v := range d.DeploymentOpts.SecretMounts {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      v.Name,
			MountPath: v.MountPath,
		})
		volumes = append(volumes, v1.Volume{
			Name: v.Name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: v.Secret,
				},
			},
		})
	}

	// Container to use
	containers := []v1.Container{
		{
			Name:            d.Name,
			Image:           d.DeploymentOpts.Image,
			ImagePullPolicy: v1.PullAlways,
			Env:             d.DeploymentOpts.EnvVars,
			Resources:       d.DeploymentOpts.ResourceReq,
			VolumeMounts:    volumeMounts,
		},
	}

	// Customize commands and arguments if any informed
	if len(d.DeploymentOpts.Command) > 0 {
		containers[0].Command = d.DeploymentOpts.Command
	}
	if len(d.DeploymentOpts.Args) > 0 {
		containers[0].Args = d.DeploymentOpts.Args
	}

	deploymentconfig := &osappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Name,
		},
		Spec: osappsv1.DeploymentConfigSpec{

			Selector: d.DeploymentOpts.Labels,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      d.Name,
					Namespace: d.Namespace.Namespace,
					Labels:    d.DeploymentOpts.Labels,
				},
				Spec: v1.PodSpec{
					Volumes:       volumes,
					Containers:    containers,
					RestartPolicy: d.DeploymentOpts.RestartPolicy,
				},
			},
			Replicas: 1,
		},
	}

	phase := frame2.Phase{
		Runner: d.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &OCPDeploymentConfig{
					Namespace:        d.Namespace,
					DeploymentConfig: deploymentconfig,
					Ctx:              ctx,
				},
			},
		},
	}
	err := phase.Run()
	if err != nil {
		return fmt.Errorf("failed to create deploymentconfig: %w", err)
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
					Validator: &OCPDeploymentConfigGet{
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

// Executes a fully specified OCP DeploymentConfig
//
// # See OCPDeploymentConfigOpts for a simpler interface
type OCPDeploymentConfig struct {
	Namespace        *base.ClusterContext
	DeploymentConfig *osappsv1.DeploymentConfig

	Result *osappsv1.DeploymentConfig
	Ctx    context.Context
}

func (d *OCPDeploymentConfig) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	var err error
	client, err := clientset.NewForConfig(d.Namespace.VanClient.RestConfig)
	if err != nil {
		return fmt.Errorf("Failed to obtain clientset")
	}
	d.Result, err = client.DeploymentConfigs(d.Namespace.Namespace).Create(
		ctx,
		d.DeploymentConfig,
		metav1.CreateOptions{},
	)

	if err != nil {
		return fmt.Errorf("Failed to create deploymentconfig: %w", err)
	}

	return nil
}

type OCPDeploymentConfigGet struct {
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	Result *osappsv1.DeploymentConfig

	frame2.Log
	frame2.DefaultRunDealer
}

func (d *OCPDeploymentConfigGet) Validate() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	client, err := clientset.NewForConfig(d.Namespace.VanClient.RestConfig)

	d.Result, err = client.DeploymentConfigs(d.Namespace.Namespace).Get(
		ctx,
		d.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("Failed to get deploymentconfig %q: %w", d.Name, err)
	}

	// TODO Change his by d.MinReplicas?
	if d.Result.Status.ReadyReplicas < 1 {
		return fmt.Errorf("DeploymentConfig %q has no ready replicas", d.Name)
	}

	return nil
}

// Wait for the named deploymentconfig to be available.  By default, it
// waits for up to two minutes, and ensures that the deployment reports
// as ready for at least 10s.
//
// That behavior can be changed using the RetryOptions field. On that
// field, the Ctx field cannot be set; if a different timeout is desired,
// set it on the Action's Ctx itself, and it will be used for the
// RetryOptions.
type OCPDeploymentConfigWait struct {
	Name      string
	Namespace *base.ClusterContext
	Ctx       context.Context

	// On this field, do not set the context.  Use the OCPDeploymentConfigWait.Ctx,
	// instead, it will be used for the underlying Retry
	RetryOptions frame2.RetryOptions
	frame2.DefaultRunDealer
	*frame2.Log
}

func (w OCPDeploymentConfigWait) Validate() error {
	if w.RetryOptions.Ctx != nil {
		panic("RetryOptions.Ctx cannot be set for OCPDeploymentConfigWait")
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
		Doc:    fmt.Sprintf("Waiting for deploymentconfig %q on ns %q", w.Name, w.Namespace.Namespace),
		MainSteps: []frame2.Step{
			{
				// TODO: stuff within functions need their runners replaced?
				ValidatorRetry: retry,
				Validator: &Function{
					Fn: func() error {
						validator := &OCPDeploymentConfigGet{
							Namespace: w.Namespace,
							Name:      w.Name,
						}
						inner1 := frame2.Phase{
							Runner: w.GetRunner(),
							Doc:    fmt.Sprintf("Get the deploymentconfig %q on ns %q", w.Name, w.Namespace.Namespace),
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
							Doc:    fmt.Sprintf("Check that the deploymentconfig %q is ready", w.Name),
							MainSteps: []frame2.Step{
								{
									Validator: &Function{
										Fn: func() error {
											if validator.Result == nil {
												return fmt.Errorf("deploymentconfig not ready: result is nil")
											}
											if validator.Result.Status.ReadyReplicas == 0 {
												return fmt.Errorf("deploymentconfig not ready: ready replicas is 0")
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

/*
 * TODO: currently, this is a copy of K8SDeployment stuff

type OCPDeploymentConfigAnnotate struct {
	Namespace   *base.ClusterContext
	Name        string
	Annotations map[string]string

	Ctx context.Context
}

func (kda OCPDeploymentConfigAnnotate) Execute() error {
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
*/

type OCPDeploymentConfigUndeploy struct {
	Name      string
	Namespace *base.ClusterContext
	Wait      time.Duration // Waits for the deployment to be gone.  Otherwise, returns as soon as the delete instruction has been issued.  If the wait lapses, return an error.

	Ctx context.Context
	frame2.DefaultRunDealer
}

func (k *OCPDeploymentConfigUndeploy) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)

	client, err := clientset.NewForConfig(k.Namespace.VanClient.RestConfig)

	err = client.DeploymentConfigs(k.Namespace.Namespace).Delete(
		ctx,
		k.Name,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return err
	}
	if k.Wait == 0 {
		return nil
	}
	phase := frame2.Phase{
		Runner: k.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Doc: "Confirm the deploymentconfig is gone",
				Validator: &OCPDeploymentConfigGet{
					Namespace: k.Namespace,
					Name:      k.Name,
					Ctx:       ctx,
				},
				ExpectError: true,
				ValidatorRetry: frame2.RetryOptions{
					Ctx:        ctx,
					Timeout:    k.Wait,
					KeepTrying: true,
				},
			},
		},
	}
	return phase.Run()
}
