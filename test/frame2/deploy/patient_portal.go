package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/validate"
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
	Topology *topology.Basic

	// This will create K8S services
	CreateServices bool

	// This will create Skupper services; if CreateServices is also
	// true, the Skupper service will be based on the K8S service.
	// Otherwise, it exposes the deployment.
	//
	// The Skupper service will use the HTTP protocol
	SkupperExpose bool

	frame2.DefaultRunDealer
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

	front_ns, err := (*p.Topology).Get(topology.Public, 1)
	if err != nil {
		return fmt.Errorf("failed to get frontend namespace")
	}

	phase := frame2.Phase{
		Runner: p.Runner,
		Doc:    "Install Patient Portal pieces",
		MainSteps: []frame2.Step{
			{
				Doc: "Install Patient Portal database",
				Modify: &PatientDatabase{
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
					Target:         front_ns,
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

	frame2.DefaultRunDealer
}

func (p PatientDatabase) Execute() error {
	ctx := frame2.ContextOrDefault(p.Ctx)

	image := p.Image
	if image == "" {
		image = "quay.io/dhashimo/patient-portal-database"
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
				Modify: &ExposeHelper{
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
		image = "quay.io/dhashimo/patient-portal-payment-processor"
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
				Modify: &ExposeHelper{
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
		image = "quay.io/dhashimo/patient-portal-frontend"
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
				Modify: &ExposeHelper{
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

type PatientValidatePayment struct {
	Namespace   *base.ClusterContext
	ServiceName string // default is payment-processor
	ServicePort int    // default is 8080
	ServicePath string // default is api/pay

	frame2.Log
	frame2.DefaultRunDealer
}

func (p PatientValidatePayment) Validate() error {
	if p.Namespace == nil {
		return fmt.Errorf("PatientValidatePayment configuration error: empty Namespace")
	}
	svc := p.ServiceName
	if svc == "" {
		svc = "payment-processor"
	}
	port := p.ServicePort
	if port == 0 {
		port = 8080
	}
	path := p.ServicePath
	if path == "" {
		path = "api/pay"
	}
	phase := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &validate.Curl{
					Namespace:   p.Namespace,
					Url:         fmt.Sprintf("http://%s:%d/%s", svc, port, path),
					Fail400Plus: true,
					Log:         p.Log,
				},
			},
		},
	}
	phase.SetLogger(p.Logger)
	return phase.Run()
}

type PatientFrontendHealth struct {
	Namespace   *base.ClusterContext
	ServiceName string // default is frontend
	ServicePort int    // default is 8080
	ServicePath string // default is api/health

	frame2.Log
	frame2.DefaultRunDealer
}

func (p PatientFrontendHealth) Validate() error {
	if p.Namespace == nil {
		return fmt.Errorf("PatientCurlFrontend configuration error: empty Namespace")
	}
	svc := p.ServiceName
	if svc == "" {
		svc = "frontend"
	}
	port := p.ServicePort
	if port == 0 {
		port = 8080
	}
	path := p.ServicePath
	if path == "" {
		path = "api/health"
	}
	phase := frame2.Phase{
		Runner: p.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &validate.Curl{
					Namespace:   p.Namespace,
					Url:         fmt.Sprintf("http://%s:%d/%s", svc, port, path),
					Fail400Plus: true,
					Log:         p.Log,
				},
			},
		},
	}
	phase.SetLogger(p.Logger)
	return phase.Run()
}

// Given a namespace with a PatientFrontend deployment, it will ping the
// DB from that deployment using pg_isready
// TODO change this to use a test helper pod, instead of the frontend
type PatientDbPing struct {
	Namespace *base.ClusterContext

	frame2.Log
	frame2.DefaultRunDealer
}

func (p PatientDbPing) Validate() error {
	phase := frame2.Phase{
		Runner: p.Runner,
		Log:    p.Log,
		Doc:    "Ping the DB",
		MainSteps: []frame2.Step{
			{
				Validator: &execute.PostgresPing{
					Namespace: p.Namespace,
					Labels:    map[string]string{"app": "frontend"},
					DbName:    "database",
					DbHost:    "database",
					Username:  "patient_portal",
					Log:       p.Log,
				},
			},
		},
	}
	return phase.Run()
}
