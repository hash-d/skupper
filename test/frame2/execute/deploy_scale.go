package execute

import (
	"log"

	"github.com/skupperproject/skupper/test/utils/base"
)

type DeployScale struct {
	Namespace      base.ClusterContextPromise
	DeploySelector // Do not populate the Namespace within the PodSelector; it will be auto-populated
	Replicas       int32
}

func (d DeployScale) Execute() error {
	log.Printf("execute.DeployScale")

	cluster, err := d.Namespace.Satisfy()

	d.DeploySelector.Namespace = d.Namespace

	err = d.DeploySelector.Execute()
	if err != nil {
		return err
	}

	deploy := d.DeploySelector.Deploy

	deploy.Spec.Replicas = &d.Replicas
	_, err = cluster.VanClient.KubeClient.AppsV1().
		Deployments(cluster.Namespace).Update(deploy)

	return err

}
