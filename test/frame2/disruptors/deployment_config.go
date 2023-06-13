package disruptors

import (
	"log"

	osappsv1 "github.com/openshift/api/apps/v1"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This disruptor will try to change any uses of K8S deployments into
// DeploymentConfig, on test code:
//
// - on actual deployments
// - on skupper expose
//
// As the name says, it will do it 'blindly'; it's useful only for very
// small and simple tests.
//
// Part of its work is to change a K8SDeploy action by a DeploymentConfig
// action; there is nothing to prevent that earlier parts of the test
// kept pointers to the old action.
type DeploymentConfigBlindly struct {
}

func (d DeploymentConfigBlindly) DisruptorEnvValue() string {
	return "DEPLOYMENTCONFIG_BLINDLY"
}

func (d DeploymentConfigBlindly) Inspect(step *frame2.Step, phase *frame2.Phase) {
	switch mod := step.Modify.(type) {

	case *execute.K8SDeployment:
		log.Printf("[D] DEPLOYMENTCONFIG_BLINDLY changed modifier %v Type to 'deploymentconfig'", mod.Deployment.Name)
		log.Printf("[D] Before: %#v", mod.Deployment)

		// K8S' Deployment has a default of 1 Replicas, while DeploymentConfigs default to zero
		replicas := mod.Deployment.Spec.Replicas
		if replicas == nil || *replicas == 0 {
			var one int32 = 1

			replicas = &one
		}

		// This ignores a whole lot of different things that could be defined on the deployment
		deploymentconfig := &osappsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mod.Deployment.Name,
				Namespace: mod.Deployment.Namespace,
				Labels:    mod.Deployment.Labels,
			},
			Spec: osappsv1.DeploymentConfigSpec{
				Selector: mod.Deployment.Labels,
				Replicas: *replicas,
				Template: &v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: mod.Deployment.Labels,
					},
					Spec: v1.PodSpec{
						Volumes:       mod.Deployment.Spec.Template.Spec.Volumes,
						Containers:    mod.Deployment.Spec.Template.Spec.Containers,
						RestartPolicy: mod.Deployment.Spec.Template.Spec.RestartPolicy,
					},
				},
			},
		}

		log.Printf("[D] After: %#v", deploymentconfig)
		step.Modify = &execute.OCPDeploymentConfig{
			Namespace:        mod.Namespace,
			DeploymentConfig: deploymentconfig,
			Ctx:              mod.Ctx,
		}
	case *execute.K8SDeploymentOpts:
		log.Printf("[D] DEPLOYMENT_CONFIGS_BLINDLY overriding K8SDeploymentOpts %q", mod.Name)
		newMod := execute.OCPDeploymentConfigOpts{
			Name:             mod.Name,
			Namespace:        mod.Namespace,
			DeploymentOpts:   mod.DeploymentOpts,
			Wait:             mod.Wait,
			Ctx:              mod.Ctx,
			DefaultRunDealer: mod.DefaultRunDealer,
		}
		step.Modify = &newMod
	case *execute.SkupperExpose:
		if mod.Type == "deployment" {
			log.Printf("[D] DEPLOYMENTCONFIG_BLINDLY overriding SkupperExpose for %q as 'deploymentconfig'", mod.Name)
			mod.Type = "deploymentconfig"
		}
	case *execute.K8SUndeploy:
		newMod := execute.OCPDeploymentConfigUndeploy(*mod)
		step.Modify = &newMod
	}

	checkValidator := func(v frame2.Validator) (frame2.Validator, bool) {
		if v == nil {
			return nil, false
		}
		switch val := v.(type) {
		case execute.K8SDeploymentWait:
			transformed := execute.OCPDeploymentConfigWait(val)
			return transformed, true
		case *execute.K8SDeploymentGet:
			transformed := execute.OCPDeploymentConfigGet{
				Namespace:        val.Namespace,
				Name:             val.Name,
				Ctx:              val.Ctx,
				DefaultRunDealer: val.DefaultRunDealer,
			}
			return &transformed, true
		}
		return nil, false
	}

	// TODO This should be recurring enough to be worthy of a helper
	if v, changed := checkValidator(step.Validator); changed {
		step.Validator = v
	}
	for i, val := range step.Validators {
		if v, changed := checkValidator(val); changed {
			step.Validators[i] = v
		}
	}

}
