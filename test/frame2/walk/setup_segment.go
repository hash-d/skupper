package walk

import (
	"context"
	"log"

	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/constants"
	"github.com/skupperproject/skupper/test/utils/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SegmentSetup struct {
	Namespace *base.ClusterContext
	Runner    *base.ClusterTestRunnerBase
}

// Right now, this is a copy of hello world's setup.  The
// idea, however, is to split it in a bunch of individual
// frame2.step and then compose it back.
func (s SegmentSetup) Execute() error {

	runner := s.Runner

	log.Printf("Segment setup: %+v", s)
	needs := base.ClusterNeeds{
		NamespaceId:     "hello-world",
		PublicClusters:  1,
		PrivateClusters: 2,
	}
	if err := runner.Validate(needs); err != nil {
		return err
	}
	log.Printf("Building the environment")
	_, err := runner.Build(needs, nil)
	if err != nil {
		return err
	}

	// getting public and private contexts
	pub, err := runner.GetPublicContext(1)
	if err != nil {
		return err
	}
	prv, err := runner.GetPrivateContext(1)
	if err != nil {
		return err
	}

	prv2, err := runner.GetPrivateContext(2)
	if err != nil {
		return err
	}

	//	// creating namespaces
	//	if err = pub.CreateNamespace(); err != nil {
	//		return err
	//	}
	//	if err = prv.CreateNamespace(); err != nil {
	//		return err
	//	}

	ctx := context.Background()
	if err = base.SetupSimplePublicPrivate(ctx, runner); err != nil {
		return err
	}
	if err = deployResources(pub, prv); err != nil {
		return err
	}
	if err = base.ConnectSimplePublicPrivate(ctx, runner); err != nil {
		return err
	}
	err = prv2.CreateNamespace()
	if err != nil {
		return err
	}

	log.Printf("Segment done")

	return nil
}

// deployResources Deploys the hello-world-frontend and hello-world-backend
// pods and validate they are available
func deployResources(pub *base.ClusterContext, prv *base.ClusterContext) error {
	frontend, _ := k8s.NewDeployment("hello-world-frontend", pub.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-frontend",
		Labels:        map[string]string{"app": "hello-world-frontend"},
		RestartPolicy: corev1.RestartPolicyAlways,
	})
	backend, _ := k8s.NewDeployment("hello-world-backend", prv.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-backend",
		Labels:        map[string]string{"app": "hello-world-backend"},
		RestartPolicy: corev1.RestartPolicyAlways,
	})

	ctx := context.Background()

	// Creating deployments
	if _, err := pub.VanClient.KubeClient.AppsV1().Deployments(pub.Namespace).Create(ctx, frontend, metav1.CreateOptions{}); err != nil {
		return err
	}
	if _, err := prv.VanClient.KubeClient.AppsV1().Deployments(prv.Namespace).Create(ctx, backend, metav1.CreateOptions{}); err != nil {
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

type SegmentTeardown struct {
	frame2.Step
	Runner *base.ClusterTestRunnerBase
}

func (s SegmentTeardown) Execute() error {
	base.TearDownSimplePublicAndPrivate(s.Runner)
	return nil
}
