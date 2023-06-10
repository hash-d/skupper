package validate

import (
	"context"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SkupperInfoContents struct {
	Images SkupperManifestContent

	HasRouter            bool
	HasServiceController bool
	HasPrometheus        bool

	RouterDeployment            *appsv1.Deployment
	ServiceControllerDeployment *appsv1.Deployment
	PrometheusDeployment        *appsv1.Deployment
}

const (
	SkupperRouterRepo = "https://github.com/skupperproject/skupper-router"
	SkupperRepo       = "https://github.com/skupperproject/skupper"
	EmptyRepo         = ""
	UnknownRepo       = "UNKNOWN"
)

// Gets various information about Skupper
// TODO: add ConfigMaps, skmanage executions
type SkupperInfo struct {
	Namespace *base.ClusterContext

	Result SkupperInfoContents

	Ctx context.Context
	frame2.DefaultRunDealer
	frame2.Log
}

func (s *SkupperInfo) Validate() error {
	ctx := frame2.ContextOrDefault(s.Ctx)

	var err error

	// Router deployment
	s.Result.RouterDeployment, err = s.Namespace.VanClient.KubeClient.AppsV1().Deployments(s.Namespace.Namespace).Get(ctx, types.TransportDeploymentName, metav1.GetOptions{})
	if err != nil {
		s.Log.Printf("failed to get deployment %q: %v", types.TransportDeploymentName, err)
	} else {
		s.Result.HasRouter = true
		for _, container := range s.Result.RouterDeployment.Spec.Template.Spec.Containers {
			switch container.Name {
			case types.TransportComponentName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: SkupperRouterRepo,
					},
				)
			case types.ConfigSyncContainerName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: SkupperRepo,
					},
				)
			default:
				s.Log.Printf("Unknown container %q in deployment %q", container.Name, s.Result.RouterDeployment.Name)
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: UnknownRepo,
					},
				)
			}
		}

	}

	// Service Controller Deployment
	s.Result.ServiceControllerDeployment, err = s.Namespace.VanClient.KubeClient.AppsV1().Deployments(s.Namespace.Namespace).Get(ctx, types.ControllerDeploymentName, metav1.GetOptions{})
	if err != nil {
		s.Log.Printf("failed to get deployment %q: %v", types.TransportDeploymentName, err)
	} else {
		s.Result.HasServiceController = true
		for _, container := range s.Result.ServiceControllerDeployment.Spec.Template.Spec.Containers {
			switch container.Name {
			case types.ControllerContainerName, types.FlowCollectorContainerName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: SkupperRepo,
					},
				)
			default:
				s.Log.Printf("Unknown container %q in deployment %q", container.Name, s.Result.RouterDeployment.Name)
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: UnknownRepo,
					},
				)
			}
		}

	}

	// Prometheus deployment
	s.Result.PrometheusDeployment, err = s.Namespace.VanClient.KubeClient.AppsV1().Deployments(s.Namespace.Namespace).Get(ctx, types.PrometheusDeploymentName, metav1.GetOptions{})
	if err != nil {
		s.Log.Printf("failed to get deployment %q: %v", types.TransportDeploymentName, err)
	} else {
		s.Result.HasPrometheus = true
		for _, container := range s.Result.PrometheusDeployment.Spec.Template.Spec.Containers {
			switch container.Name {
			case types.PrometheusContainerName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: EmptyRepo,
					},
				)
			default:
				s.Log.Printf("Unknown container %q in deployment %q", container.Name, s.Result.RouterDeployment.Name)
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: UnknownRepo,
					},
				)
			}
		}

	}

	return nil

}
