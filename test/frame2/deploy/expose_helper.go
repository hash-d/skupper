package deploy

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
	apiv1 "k8s.io/api/core/v1"
)

// ExposeHelper creates K8S services and/or Skupper services for a deployment
//
// As its name implies, it's just a helper.  Several 'deploy' pieces would repeat
// this code, so it's been extracted for reuse
type ExposeHelper struct {
	Target *base.ClusterContext

	// This will create K8S services
	CreateServices bool

	// This will create Skupper services; if CreateServices is also
	// true, the Skupper service will be based on the K8S service.
	// Otherwise, it exposes the deployment.
	//
	// The Skupper service will use the HTTP protocol
	SkupperExpose bool

	ServiceName   string
	ServicePorts  []int
	ServiceLabels map[string]string
	ServiceType   apiv1.ServiceType

	Protocol string

	frame2.DefaultRunDealer

	//Ctx context.Context
}

func (e ExposeHelper) Execute() error {
	//ctx := frame2.ContextOrDefault(e.Ctx)

	ports32 := make([]int32, len(e.ServicePorts))

	for i, p := range e.ServicePorts {
		ports32[i] = int32(p)
	}

	phase := frame2.Phase{
		Runner: e.Runner,
		MainSteps: []frame2.Step{
			{

				Doc: fmt.Sprintf("Creating a local service for %q", e.ServiceName),
				Modify: &execute.K8SServiceCreate{
					Namespace: e.Target,
					Name:      e.ServiceName,
					Labels:    e.ServiceLabels,
					Selector:  e.ServiceLabels,
					Ports:     ports32,
					Type:      e.ServiceType,
				},
				SkipWhen: !e.CreateServices,
			}, {
				Doc: "Exposing the local service via Skupper",
				Modify: &execute.SkupperExpose{
					Namespace: e.Target,
					Type:      "service",
					Name:      e.ServiceName,
					Protocol:  e.Protocol,
				},
				SkipWhen: !e.CreateServices || !e.SkupperExpose,
			}, {
				Doc: "Exposing the deployment via Skupper",
				Modify: &execute.SkupperExpose{
					Namespace: e.Target,
					Ports:     e.ServicePorts,
					Type:      "deployment",
					Name:      e.ServiceName,
					Protocol:  e.Protocol,
				},
				SkipWhen: e.CreateServices || !e.SkupperExpose,
			},
		},
	}
	return phase.Run()
}
