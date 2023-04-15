package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	v1 "k8s.io/api/core/v1"
)

// A full deployment of Patient Portal
//
// For fine tuned configuration, use the individual PatientDatabase, PatientFrontend
// and PatientPayment components
type PatientPortal struct {
	Runner   *frame2.Run
	Topology *topology.Basic

	// This will create K8S services
	CreateServices bool

	// This will create Skupper services; if CreateServices is also
	// true, the Skupper service will be based on the K8S service.
	// Otherwise, it exposes the deployment.
	//
	// The Skupper service will use the HTTP protocol
	SkupperExpose bool
}

func (p PatientPortal) Execute() error {

	//	pub, err := (*p.Topology).Get(topology.Public, 1)
	//	if err != nil {
	//		return fmt.Errorf("failed to get public-1")
	//	}
	db_ns, err := (*p.Topology).Get(topology.Private, 1)
	if err != nil {
		return fmt.Errorf("failed to get database namespace")
	}

	phase := frame2.Phase{
		Runner: p.Runner,
		Doc:    "Install Patient Portal pieces",
		MainSteps: []frame2.Step{
			{
				Doc: "Install Patient Portal database",
				Modify: &PatientDatabase{
					Runner:         p.Runner,
					Target:         db_ns,
					CreateServices: p.CreateServices,
					SkupperExpose:  p.SkupperExpose,
				},
			},
		},
	}
	return phase.Run()

}

// Deploys the patient database from
//
// quay.io/skupper/patient-portal-database
type PatientDatabase struct {
	Runner *frame2.Run
	Target *base.ClusterContext

	Image string // default quay.io/skupper/patient-portal-database

	// This will create K8S services
	CreateServices bool

	// This will create Skupper services; if CreateServices is also
	// true, the Skupper service will be based on the K8S service.
	// Otherwise, it exposes the deployment.
	//
	// The Skupper service will use the HTTP protocol
	SkupperExpose bool

	Ctx context.Context
}

func (p PatientDatabase) Execute() error {
	ctx := frame2.ContextOrDefault(p.Ctx)

	image := p.Image
	if image == "" {
		image = "quay.io/skupper/patient-portal-database"
	}

	labels := map[string]string{"app": "patient-portal-database"}

	phase := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing patient-portal-database",
				Modify: &execute.K8SDeploymentOpts{
					Name:      "database",
					Namespace: p.Target,
					DeploymentOpts: k8s.DeploymentOpts{
						Image:         image,
						Labels:        labels,
						RestartPolicy: v1.RestartPolicyAlways,
					},
					Wait: time.Minute * 2,
					Ctx:  ctx,
				},
			}, {
				Doc: "Creating a local service for patient-portal-database",
				Modify: &execute.K8SServiceCreate{
					Namespace: p.Target,
					Name:      "database",
					Labels:    labels,
					Ports:     []int32{5432},
				},
				SkipWhen: !p.CreateServices,
			}, {
				Doc: "Exposing the local service via Skupper",
				Modify: &execute.SkupperExpose{
					Runner:    p.Runner,
					Namespace: p.Target,
					Type:      "service",
					Name:      "database",
					Protocol:  "tcp",
				},
				SkipWhen: !p.CreateServices || !p.SkupperExpose,
			}, {
				Doc: "Exposing the deployment via Skupper",
				Modify: &execute.SkupperExpose{
					Runner:    p.Runner,
					Namespace: p.Target,
					Ports:     []int{5432},
					Type:      "deployment",
					Name:      "database",
					Protocol:  "tcp",
				},
				SkipWhen: p.CreateServices || !p.SkupperExpose,
			},
		},
	}
	return phase.Run()
}
