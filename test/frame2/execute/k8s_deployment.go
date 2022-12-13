package execute

import (
	"fmt"

	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	v1 "k8s.io/api/apps/v1"
)

// This simply makes a request to k8s.NewDeployment
//
// See K8SDeployment for a more complete interface
type K8SDeploymentOpts struct {
	Name           string
	Namespace      *base.ClusterContextPromise
	DeploymentOpts k8s.DeploymentOpts

	Result *v1.Deployment
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
	Deployment *v1.Deployment

	Result *v1.Deployment
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
