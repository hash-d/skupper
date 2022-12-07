package execute

import (
	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/utils/base"
	v1 "k8s.io/api/apps/v1"
)

type DeploySelector struct {
	Namespace base.ClusterContextPromise
	Name      string

	// Return value
	Deploy *v1.Deployment
}

func (d *DeploySelector) Execute() error {
	cluster, err := d.Namespace.Satisfy()

	if err != nil {
		return err
	}

	deploy, err := kube.GetDeployment(d.Name, cluster.Namespace, cluster.VanClient.KubeClient)
	d.Deploy = deploy
	if err != nil {
		return err
	}
	return nil
}
