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
	corev1 "k8s.io/api/core/v1"
)

// A full deployment of Patient Portal
//
// For fine tuned configuration, use the individual PatientDatabase, PatientFrontend
// and PatientPayment components
//
// https://github.com/skupperproject/skupper-example-patient-portal/
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

	payment_ns, err := (*p.Topology).Get(topology.Private, 2)
	if err != nil {
		return fmt.Errorf("failed to get payment processor namespace")
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
			}, {
				Doc: "Install Patient Portal Payment Processor service",
				Modify: &PatientPayment{
					Runner:         p.Runner,
					Target:         payment_ns,
					CreateServices: p.CreateServices,
					SkupperExpose:  p.SkupperExpose,
				},
			}, {
				Doc: "Install Patient Portal Payment Frontend",
				Modify: &PatientFrontend{
					Runner:         p.Runner,
					Target:         payment_ns,
					CreateServices: true,
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
						RestartPolicy: corev1.RestartPolicyAlways,
					},
					Wait: time.Minute * 2,
					Ctx:  ctx,
				},
			}, {
				Doc: "Creating services, as required",
				Modify: ExposeHelper{
					Runner:         p.Runner,
					Target:         p.Target,
					CreateServices: p.CreateServices,
					SkupperExpose:  p.SkupperExpose,
					ServiceName:    "database",
					ServicePorts:   []int{5432},
					ServiceLabels:  labels,
					Protocol:       "tcp",
				},
			},
		},
	}
	return phase.Run()
}

// Deploys the patient payment processor
type PatientPayment struct {
	Runner *frame2.Run
	Target *base.ClusterContext

	Image string // default quay.io/skupper/patient-portal-payment-processor

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

func (p PatientPayment) Execute() error {
	ctx := frame2.ContextOrDefault(p.Ctx)

	image := p.Image
	if image == "" {
		image = "quay.io/skupper/patient-portal-payment-processor"
	}

	labels := map[string]string{"app": "patient-portal-payment"}

	phase := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing patient-portal-payment-processor",
				Modify: &execute.K8SDeploymentOpts{
					Name:      "payment-processor",
					Namespace: p.Target,
					DeploymentOpts: k8s.DeploymentOpts{
						Image:         image,
						Labels:        labels,
						RestartPolicy: corev1.RestartPolicyAlways,
					},
					Wait: time.Minute * 2,
					Ctx:  ctx,
				},
			}, {
				Doc: "Creating services, as required",
				Modify: ExposeHelper{
					Runner:         p.Runner,
					Target:         p.Target,
					CreateServices: p.CreateServices,
					SkupperExpose:  p.SkupperExpose,
					ServiceName:    "payment-processor",
					ServicePorts:   []int{8080},
					ServiceLabels:  labels,
					Protocol:       "tcp",
				},
			},
		},
	}
	return phase.Run()
}

// Deploys the patient frontend
type PatientFrontend struct {
	Runner *frame2.Run
	Target *base.ClusterContext

	Image string // quay.io/skupper/patient-portal-frontend

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

func (p PatientFrontend) Execute() error {
	ctx := frame2.ContextOrDefault(p.Ctx)

	image := p.Image
	if image == "" {
		image = "quay.io/skupper/patient-portal-frontend"
	}

	labels := map[string]string{"app": "frontend"}

	phase := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing patient-portal-frontend",
				Modify: &execute.K8SDeploymentOpts{
					Name:      "frontend",
					Namespace: p.Target,
					DeploymentOpts: k8s.DeploymentOpts{
						Image:         image,
						Labels:        labels,
						RestartPolicy: corev1.RestartPolicyAlways,
						EnvVars: []corev1.EnvVar{
							{
								Name:  "DATABASE_SERVICE_HOST",
								Value: "database",
							}, {
								Name:  "DATABASE_SERVICE_PORT",
								Value: "5432",
							}, {
								Name:  "PAYMENT_PROCESSOR_SERVICE_HOST",
								Value: "payment-processor",
							}, {
								Name:  "PAYMENT_PROCESSOR_SERVICE_PORT",
								Value: "8080",
							},
						},
					},
					Wait: time.Minute * 2,
					Ctx:  ctx,
				},
			}, {
				Doc: "Creating services, as required",
				Modify: ExposeHelper{
					Runner:         p.Runner,
					Target:         p.Target,
					CreateServices: true,
					SkupperExpose:  false,
					ServiceName:    "frontend",
					ServicePorts:   []int{8080},
					ServiceLabels:  labels,
					ServiceType:    corev1.ServiceTypeLoadBalancer,
					Protocol:       "tcp",
				},
			},
		},
	}
	return phase.Run()
}
