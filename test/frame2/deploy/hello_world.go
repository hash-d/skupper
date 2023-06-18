package deploy

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
	v1 "k8s.io/api/core/v1"
)

// Deploys HelloWorld; frontend on pub1, backend on prv1
type HelloWorld struct {
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
		Runner: hw.Runner,
		Doc:    "Install Hello World front and back ends",
		MainSteps: []frame2.Step{
			{
				Doc: "Install Hello World frontend",
				Modify: &HelloWorldFrontend{
					Target:         pub,
					CreateServices: hw.CreateServices,
					SkupperExpose:  hw.SkupperExpose,
				},
			}, {
				Doc: "Install Hello World backend",
				Modify: &HelloWorldBackend{
					Target:         prv,
					CreateServices: hw.CreateServices,
					SkupperExpose:  hw.SkupperExpose,
				},
			},
		},
	}
	return phase.Run()
}

type HelloWorldBackend struct {
	Target         *base.ClusterContext
	CreateServices bool
	SkupperExpose  bool
	Protocol       string // This will default to http if not specified

	Ctx context.Context
	frame2.DefaultRunDealer
}

func (h *HelloWorldBackend) Execute() error {

	ctx := frame2.ContextOrDefault(h.Ctx)

	proto := h.Protocol
	if proto == "" {
		proto = "http"
	}

	labels := map[string]string{"app": "hello-world-backend"}

	d, err := k8s.NewDeployment("hello-world-backend", h.Target.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-backend",
		Labels:        labels,
		RestartPolicy: v1.RestartPolicyAlways,
	})
	if err != nil {
		return fmt.Errorf("HelloWorldBackend: failed to deploy: %w", err)
	}

	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing hello-world-backend",
				Modify: &execute.K8SDeployment{
					Namespace:  h.Target,
					Deployment: d,
					Ctx:        ctx,
				},
			}, {
				Doc: "Creating a local service for hello-world-backend",
				Modify: &execute.K8SServiceCreate{
					Namespace: h.Target,
					Name:      "hello-world-backend",
					Labels:    labels,
					Ports:     []int32{8080},
				},
				SkipWhen: !h.CreateServices,
			}, {
				Doc: "Exposing the local service via Skupper",
				Modify: &execute.SkupperExpose{
					Namespace: h.Target,
					Type:      "service",
					Name:      "hello-world-backend",
					Protocol:  proto,
				},
				SkipWhen: !h.CreateServices || !h.SkupperExpose,
			}, {
				Doc: "Exposing the deployment via Skupper",
				Modify: &execute.SkupperExpose{
					Namespace: h.Target,
					Ports:     []int{8080},
					Type:      "deployment",
					Name:      "hello-world-backend",
					Protocol:  proto,
				},
				SkipWhen: h.CreateServices || !h.SkupperExpose,
				Validator: execute.K8SDeploymentWait{
					Namespace: h.Target,
					Name:      "hello-world-backend",
				},
			},
		},
	}
	return phase.Run()
}

type HelloWorldFrontend struct {
	Target         *base.ClusterContext
	CreateServices bool
	SkupperExpose  bool
	Protocol       string // This will default to http if not specified

	Ctx context.Context

	frame2.DefaultRunDealer
}

func (h *HelloWorldFrontend) Execute() error {

	ctx := frame2.ContextOrDefault(h.Ctx)

	proto := h.Protocol
	if proto == "" {
		proto = "http"
	}

	labels := map[string]string{"app": "hello-world-frontend"}

	d, err := k8s.NewDeployment("hello-world-frontend", h.Target.Namespace, k8s.DeploymentOpts{
		Image:         "quay.io/skupper/hello-world-frontend",
		Labels:        labels,
		RestartPolicy: v1.RestartPolicyAlways,
	})
	if err != nil {
		return fmt.Errorf("HelloWorldFrontend: failed to deploy: %w", err)
	}

	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing hello-world-frontend",
				Modify: &execute.K8SDeployment{
					Namespace:  h.Target,
					Deployment: d,
					Ctx:        ctx,
				},
			}, {
				Doc: "Creating a local service for hello-world-frontend",
				Modify: &execute.K8SServiceCreate{
					Namespace: h.Target,
					Name:      "hello-world-frontend",
					Labels:    labels,
					Ports:     []int32{8080},
				},
				SkipWhen: !h.CreateServices,
			}, {
				Doc: "Exposing the local service via Skupper",
				Modify: &execute.SkupperExpose{
					Namespace: h.Target,
					Type:      "service",
					Name:      "hello-world-frontend",
					Protocol:  proto,
				},
				SkipWhen: !h.CreateServices || !h.SkupperExpose,
			}, {
				Doc: "Exposing the deployment via Skupper",
				Modify: &execute.SkupperExpose{
					Namespace: h.Target,
					Ports:     []int{8080},
					Type:      "deployment",
					Name:      "hello-world-frontend",
					Protocol:  proto,
				},
				Validator: execute.K8SDeploymentWait{
					Namespace: h.Target,
					Name:      "hello-world-frontend",
				},
				SkipWhen: h.CreateServices || !h.SkupperExpose,
			},
		},
	}
	return phase.Run()
}

// Validates a Hello World deployment by Curl from the given Namespace.
//
// The individual validaators (front and back) may be configured, but generally do not need to;
// they'll use the default values.
type HelloWorldValidate struct {
	Namespace               *base.ClusterContext
	HelloWorldValidateFront HelloWorldValidateFront
	HelloWorldValidateBack  HelloWorldValidateBack

	frame2.Log
	frame2.DefaultRunDealer
}

func (h HelloWorldValidate) Validate() error {
	if h.Namespace == nil {
		return fmt.Errorf("HelloWorldValidate configuration error: empty Namespace")
	}

	if h.HelloWorldValidateFront.Namespace == nil {
		h.HelloWorldValidateFront.Namespace = h.Namespace
	}
	if h.HelloWorldValidateFront.Runner == nil {
		h.HelloWorldValidateFront.Runner = h.Runner
	}
	h.HelloWorldValidateFront.OrSetLogger(h.Log.GetLogger())

	if h.HelloWorldValidateBack.Namespace == nil {
		h.HelloWorldValidateBack.Namespace = h.Namespace
	}
	if h.HelloWorldValidateBack.Runner == nil {
		h.HelloWorldValidateBack.Runner = h.Runner
	}
	h.HelloWorldValidateBack.OrSetLogger(h.Log.GetLogger())

	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Validators: []frame2.Validator{
					&h.HelloWorldValidateFront,
					&h.HelloWorldValidateBack,
				},
			},
		},
	}
	phase.OrSetLogger(h.Logger)
	return phase.Run()
}

type HelloWorldValidateFront struct {
	Namespace   *base.ClusterContext
	ServiceName string // default is hello-world-frontend
	ServicePort int    // default is 8080

	frame2.Log
	frame2.DefaultRunDealer
}

func (h HelloWorldValidateFront) Validate() error {
	if h.Namespace == nil {
		return fmt.Errorf("HelloWorldValidateFront configuration error: empty Namespace")
	}
	svc := h.ServiceName
	if svc == "" {
		svc = "hello-world-frontend"
	}
	port := h.ServicePort
	if port == 0 {
		port = 8080
	}
	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &validate.Curl{
					Namespace:   h.Namespace,
					Url:         fmt.Sprintf("http://%s:%d", svc, port),
					Fail400Plus: true,
					Log:         h.Log,
				},
			},
		},
	}
	phase.SetLogger(h.Logger)
	return phase.Run()
}

type HelloWorldValidateBack struct {
	Namespace   *base.ClusterContext
	ServiceName string // default is hello-world-backend
	ServicePort int    // default is 8080
	ServicePath string // default is api/hello

	frame2.Log
	frame2.DefaultRunDealer
}

func (h HelloWorldValidateBack) Validate() error {
	if h.Namespace == nil {
		return fmt.Errorf("HelloWorldValidateBack configuration error: empty Namespace")
	}
	svc := h.ServiceName
	if svc == "" {
		svc = "hello-world-backend"
	}
	port := h.ServicePort
	if port == 0 {
		port = 8080
	}
	path := h.ServicePath
	if path == "" {
		path = "api/hello"
	}
	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &validate.Curl{
					Namespace:   h.Namespace,
					Url:         fmt.Sprintf("http://%s:%d/%s", svc, port, path),
					Fail400Plus: true,
					Log:         h.Log,
				},
			},
		},
	}
	phase.SetLogger(h.Logger)
	return phase.Run()
}
