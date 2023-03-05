package deploy

import (
	"fmt"

	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/constants"
	"github.com/skupperproject/skupper/test/utils/k8s"
	v1 "k8s.io/api/core/v1"
)

type HelloWorld struct {
	Runner   *frame2.Run
	Topology *topology.Basic
}

// TODO Replace this by individual components
// deployResources Deploys the hello-world-frontend and hello-world-backend
// pods and validate they are available
func (hw HelloWorld) Execute() error {

	phase := frame2.Phase{
		Runner: hw.Runner,
		Setup: []frame2.Step{
			{
				Modify: *hw.Topology,
			},
		},
	}
	err := phase.Run()
	if err != nil {
		return fmt.Errorf("deploy.HelloWorld: failed to execute topology: %w", err)
	}

	pub, err := (*hw.Topology).Get(topology.Public, 1)
	prv, err := (*hw.Topology).Get(topology.Private, 1)

	frontend, _ := k8s.NewDeployment("hello-world-frontend", pub.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-frontend",
		Labels:        map[string]string{"app": "hello-world-frontend"},
		RestartPolicy: v1.RestartPolicyAlways,
	})
	backend, _ := k8s.NewDeployment("hello-world-backend", prv.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-backend",
		Labels:        map[string]string{"app": "hello-world-backend"},
		RestartPolicy: v1.RestartPolicyAlways,
	})

	// Creating deployments
	if _, err := pub.VanClient.KubeClient.AppsV1().Deployments(pub.Namespace).Create(frontend); err != nil {
		return err
	}
	if _, err := prv.VanClient.KubeClient.AppsV1().Deployments(prv.Namespace).Create(backend); err != nil {
		return err
	}

	// Waiting for deployments to be ready
	if _, err := kube.WaitDeploymentReady("hello-world-frontend", pub.Namespace, pub.VanClient.KubeClient, constants.ImagePullingAndResourceCreationTimeout, constants.DefaultTick); err != nil {
		return err
	}
	if _, err := kube.WaitDeploymentReady("hello-world-backend", prv.Namespace, prv.VanClient.KubeClient, constants.ImagePullingAndResourceCreationTimeout, constants.DefaultTick); err != nil {
		return err
	}

	return nil
}
