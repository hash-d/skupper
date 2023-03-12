package deploy

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/constants"
	"github.com/skupperproject/skupper/test/utils/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploys HelloWorld; frontend on pub1, backend on prv1
type HelloWorld struct {
	Runner   *frame2.Run
	Topology *topology.Basic

	// This will create K8S services
	CreateServices bool

	// This will create Skupper services; if CreateServices is also
	// true, the Skupper service will be based on the K8S service.
	// Otherwise, it exposes the deployment
	SkupperExpose bool
}

// Deploys the hello-world-frontend pod on pub1 and hello-world-backend pod on
// prv1, and validate they are available
func (hw HelloWorld) Execute() error {

	pub, err := (*hw.Topology).Get(topology.Public, 1)
	if err != nil {
		return fmt.Errorf("failed to get public-1")
	}
	prv, err := (*hw.Topology).Get(topology.Private, 1)
	if err != nil {
		return fmt.Errorf("failed to get private-1")
	}

	phase := frame2.Phase{
		MainSteps: []frame2.Step{
			{
				Modify: &HelloWorldFrontend{
					Runner:         hw.Runner,
					Target:         pub,
					CreateServices: hw.CreateServices,
				},
			}, {
				Modify: &HelloWorldBackend{
					Runner:         hw.Runner,
					Target:         prv,
					CreateServices: hw.CreateServices,
				},
			},
		},
	}
	phase.Run()

	return nil
}

type HelloWorldBackend struct {
	Runner         *frame2.Run
	Target         *base.ClusterContext
	CreateServices bool
	Ctx            context.Context
}

func (h *HelloWorldBackend) Execute() error {
	ctx := frame2.ContextOrDefault(h.Ctx)
	backend, _ := k8s.NewDeployment("hello-world-backend", h.Target.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-backend",
		Labels:        map[string]string{"app": "hello-world-backend"},
		RestartPolicy: v1.RestartPolicyAlways,
	})

	// Creating deployments
	if _, err := h.Target.VanClient.KubeClient.AppsV1().Deployments(h.Target.Namespace).Create(ctx, backend, metav1.CreateOptions{}); err != nil {
		return err
	}

	// Waiting for deployments to be ready
	if _, err := kube.WaitDeploymentReady("hello-world-backend", h.Target.Namespace, h.Target.VanClient.KubeClient, constants.ImagePullingAndResourceCreationTimeout, constants.DefaultTick); err != nil {
		return err
	}

	return nil
}

type HelloWorldFrontend struct {
	Runner         *frame2.Run
	Target         *base.ClusterContext
	CreateServices bool

	Ctx context.Context
}

func (h *HelloWorldFrontend) Execute() error {
	ctx := frame2.ContextOrDefault(h.Ctx)

	d, err := k8s.NewDeployment("hello-world-frontend", h.Target.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-frontend",
		Labels:        map[string]string{"app": "hello-world-frontend"},
		RestartPolicy: v1.RestartPolicyAlways,
	})
	if err != nil {
		return fmt.Errorf("HelloWorldFrontend: failed to deploy: %w", err)
	}
	if _, err := h.Target.VanClient.KubeClient.AppsV1().Deployments(h.Target.Namespace).Create(ctx, d, metav1.CreateOptions{}); err != nil {
		return err
	}
	if _, err := kube.WaitDeploymentReady("hello-world-frontend", h.Target.Namespace, h.Target.VanClient.KubeClient, constants.ImagePullingAndResourceCreationTimeout, constants.DefaultTick); err != nil {
		return err
	}
	return nil
}
