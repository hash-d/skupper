package main

import (
	"fmt"
	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/client"
	"gotest.tools/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"testing"
)

func TestGetServiceDefinitionFromAnnotatedDeployment(t *testing.T) {

	// Types to compose test table
	type result struct {
		service types.ServiceInterface
		success bool
	}

	type test struct {
		name       string
		deployment *v1.Deployment
		expected   result
	}

	// Mock VanClient
	const NS = "test"
	vanClient := &client.VanClient{
		Namespace:  NS,
		KubeClient: fake.NewSimpleClientset(),
	}

	dm := &DefinitionMonitor{
		vanClient: vanClient,
	}

	// Help preparing sample deployments to compose test table
	newDeployment := func(name string, proxyAnnotationProtocol string, containerPortAnnotation string, addressAnnotation string, containerPort int, labels map[string]string) *v1.Deployment {
		// Add port to container if > 0
		containerPorts := []corev1.ContainerPort{}
		if containerPort > 0 {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				Name:          "port",
				ContainerPort: int32(containerPort),
			})
		}

		// Prepare the container
		depContainers := []corev1.Container{{
			Name:  "container",
			Ports: containerPorts,
		}}

		// Deployment annotations
		annotations := map[string]string{}
		if proxyAnnotationProtocol != "" {
			annotations[types.ProxyQualifier] = proxyAnnotationProtocol
		}
		if containerPortAnnotation != "" {
			annotations[types.PortQualifier] = containerPortAnnotation
		}
		if addressAnnotation != "" {
			annotations[types.AddressQualifier] = addressAnnotation
		}

		// Only initialize the selector pointer if a label has been provided
		var selector *metav1.LabelSelector
		if len(labels) > 0 {
			selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
		}
		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   NS,
				Annotations: annotations,
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: depContainers,
					},
				},
				Selector: selector,
			},
		}

	}

	// labels to use while preparing the test table
	selectorWithLabels := map[string]string{"label1": "value1"}
	selectorWithoutLabels := map[string]string{}

	// test table below is meant to cover getServiceDefinitionFromAnnotatedDeployment()
	testTable := []test{
		{"no-proxy-annotation", newDeployment("dep1", "", "", "", 8080, selectorWithLabels), result{
			service: types.ServiceInterface{},
			success: false,
		}},
		{"http-port-annotation-no-address", newDeployment("dep1", "http", "81", "", 8080, selectorWithLabels), result{
			service: types.ServiceInterface{
				Address:  "dep1",
				Protocol: "http",
				Port:     81,
				Targets:  []types.ServiceInterfaceTarget{{Name: "dep1", Selector: "label1=value1"}},
				Origin:   "annotation",
			},
			success: true,
		}},
		{"http-port-annotation-no-addess-without-selector", newDeployment("dep1", "http", "81", "", 8080, selectorWithoutLabels), result{
			service: types.ServiceInterface{
				Address:  "dep1",
				Protocol: "http",
				Port:     81,
				Targets:  []types.ServiceInterfaceTarget{{Name: "dep1", Selector: ""}},
				Origin:   "annotation",
			},
			success: true,
		}},
		{"http-port-container-no-address", newDeployment("dep1", "http", "", "", 8080, selectorWithLabels), result{
			service: types.ServiceInterface{
				Address:  "dep1",
				Protocol: "http",
				Port:     8080,
				Targets:  []types.ServiceInterfaceTarget{{Name: "dep1", Selector: "label1=value1"}},
				Origin:   "annotation",
			},
			success: true,
		}},
		{"http-no-port-no-address", newDeployment("dep1", "http", "", "", 0, selectorWithLabels), result{
			service: types.ServiceInterface{
				Address:  "dep1",
				Protocol: "http",
				Port:     80,
				Targets:  []types.ServiceInterfaceTarget{{Name: "dep1", Selector: "label1=value1"}},
				Origin:   "annotation",
			},
			success: true,
		}},
		{"http-no-port-with-address", newDeployment("dep1", "http", "", "address1", 0, selectorWithLabels), result{
			service: types.ServiceInterface{
				Address:  "address1",
				Protocol: "http",
				Port:     80,
				Targets:  []types.ServiceInterfaceTarget{{Name: "dep1", Selector: "label1=value1"}},
				Origin:   "annotation",
			},
			success: true,
		}},
		{"tcp-invalid-port-no-address", newDeployment("dep1", "tcp", "invalid", "", 0, selectorWithLabels), result{
			service: types.ServiceInterface{},
			success: false,
		}},
	}

	// Iterating through the test table
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			service, success := dm.getServiceDefinitionFromAnnotatedDeployment(test.deployment)
			// Validating returned service
			assert.Equal(t, test.expected.service.Port, service.Port)
			assert.Equal(t, test.expected.service.Protocol, service.Protocol)
			assert.Equal(t, test.expected.service.Address, service.Address)
			assert.Equal(t, len(test.expected.service.Targets), len(service.Targets))
			if len(test.expected.service.Targets) > 0 {
				assert.Equal(t, test.expected.service.Targets[0].Name, service.Targets[0].Name)
				assert.Equal(t, test.expected.service.Targets[0].Selector, service.Targets[0].Selector)
			}
			assert.Equal(t, test.expected.service.Origin, service.Origin)
			// Validating overall result
			assert.Equal(t, success, test.expected.success)
		})
	}

}

func TestGetServiceDefinitionFromAnnotatedService(t *testing.T) {

	// Types to compose test table
	type result struct {
		service types.ServiceInterface
		success bool
	}

	type test struct {
		name     string
		service  *corev1.Service
		expected result
	}

	// Mock VanClient
	const NS = "test"
	vanClient := &client.VanClient{
		Namespace:  NS,
		KubeClient: fake.NewSimpleClientset(),
	}

	dm := &DefinitionMonitor{
		vanClient: vanClient,
	}

	// Helper used to prepare test table
	annotatedService := func(name string, proxyAnnotationProtocol string, addressAnnotation string, targetAnnotation string, labels map[string]string, targetPort int, ports ...int) *corev1.Service {

		annotations := map[string]string{}
		if proxyAnnotationProtocol != "" {
			annotations[types.ProxyQualifier] = proxyAnnotationProtocol
		}
		if addressAnnotation != "" {
			annotations[types.AddressQualifier] = addressAnnotation
		}
		if targetAnnotation != "" {
			annotations[types.TargetServiceQualifier] = targetAnnotation
		}

		// Only initialize the selector pointer if a label has been provided
		var selectors map[string]string
		if len(labels) > 0 {
			selectors = labels
		}

		// Only set ports, if at least one provided
		var svcPorts []corev1.ServicePort
		if len(ports) > 0 {
			for i, port := range ports {
				svcPorts = append(svcPorts, corev1.ServicePort{
					Name:       fmt.Sprintf("port%d", i),
					Port:       int32(port),
					TargetPort: intstr.FromInt(targetPort),
				})
			}
		}

		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Ports:    svcPorts,
				Selector: selectors,
			},
		}

	}

	// Create fake target services
	var err error
	// good path with target service providing port
	_, err = vanClient.KubeClient.CoreV1().Services(NS).Create(annotatedService("targetsvc", "", "", "", nil, 0, 8888))
	assert.NilError(t, err)
	// this is used to test case when protocol is http but target service does not provide a port, so it uses 80
	_, err = vanClient.KubeClient.CoreV1().Services(NS).Create(annotatedService("targetsvcnoport", "", "", "", nil, 0))
	assert.NilError(t, err)

	// Mock error when trying to get info for badtargetsvc
	vanClient.KubeClient.(*fake.Clientset).Fake.PrependReactor("get", "services", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		name := action.(k8stesting.GetAction).GetName()
		if name == "badtargetsvc" {
			return true, nil, fmt.Errorf("fake error has occurred")
		}
		return false, nil, nil
	})

	testTable := []test{
		{"no-proxy", annotatedService("", "", "", "", nil, 0), result{
			service: types.ServiceInterface{},
			success: false,
		}},
		{"no-target-no-selector", annotatedService("svc", "http", "", "", nil, 0), result{
			service: types.ServiceInterface{
				Address:  "svc",
				Protocol: "http",
				Port:     0,
			},
			success: false,
		}},
		{"http-8080-targetsvc-8888", annotatedService("svc", "http", "address", "targetsvc", nil, 0, 8080), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "http",
				Port:     8080,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:       "targetsvc",
						Selector:   "",
						TargetPort: 8888,
						Service:    "targetsvc",
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
		{"http-80-targetsvcnoport", annotatedService("svc", "http", "address", "targetsvcnoport", nil, 0), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "http",
				Port:     80,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:       "targetsvcnoport",
						Selector:   "",
						TargetPort: 0,
						Service:    "targetsvcnoport",
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
		{"tcp-noport-targetsvcnoport", annotatedService("svc", "tcp", "address", "targetsvcnoport", nil, 0), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "tcp",
				Port:     0,
			},
			success: false,
		}},
		{"tcp-noport-targetsvc-8888", annotatedService("svc", "tcp", "address", "targetsvc", nil, 0), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "tcp",
				Port:     8888,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:       "targetsvc",
						Selector:   "",
						TargetPort: 0,
						Service:    "targetsvc",
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
		{"bad-target-service", annotatedService("svc", "http", "address", "badtargetsvc", nil, 0, 8080), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "http",
				Port:     8080,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:       "badtargetsvc",
						Selector:   "",
						TargetPort: 0,
						Service:    "badtargetsvc",
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
		{"tcp-noport-targetsvc-8888", annotatedService("svc", "tcp", "address", "targetsvc", nil, 0), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "tcp",
				Port:     8888,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:       "targetsvc",
						Selector:   "",
						TargetPort: 0,
						Service:    "targetsvc",
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
		{"tcp-noport-selector", annotatedService("svc", "tcp", "address", "",
			map[string]string{"label1": "value1"}, 0), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "tcp",
				Port:     0,
			},
			success: false,
		}},
		{"http-noport-selector", annotatedService("svc", "http", "address", "",
			map[string]string{"label1": "value1"}, 0), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "http",
				Port:     80,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:     "svc",
						Selector: "label1=value1",
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
		{"http-8080-selector", annotatedService("svc", "http", "address", "",
			map[string]string{"label1": "value1"}, 8888, 8080), result{
			service: types.ServiceInterface{
				Address:  "address",
				Protocol: "http",
				Port:     8080,
				Targets: []types.ServiceInterfaceTarget{
					{
						Name:       "svc",
						Selector:   "label1=value1",
						TargetPort: 8888,
					},
				},
				Origin: "annotation",
			},
			success: true,
		}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			service, success := dm.getServiceDefinitionFromAnnotatedService(test.service)
			assert.Equal(t, test.expected.success, success)
			assert.Equal(t, test.expected.service.Port, service.Port)
			assert.Equal(t, test.expected.service.Address, service.Address)
			assert.Equal(t, len(test.expected.service.Targets), len(service.Targets))
			if len(service.Targets) > 0 {
				assert.Equal(t, test.expected.service.Targets[0].Name, service.Targets[0].Name)
				assert.Equal(t, test.expected.service.Targets[0].Service, service.Targets[0].Service)
				assert.Equal(t, test.expected.service.Targets[0].TargetPort, service.Targets[0].TargetPort)
				assert.Equal(t, test.expected.service.Targets[0].Selector, service.Targets[0].Selector)
			}
			assert.Equal(t, test.expected.service.Origin, service.Origin)
		})
	}

}

func TestDeducePort(t *testing.T) {

	// Helps generating the test table
	type test struct {
		name         string
		deployment   *v1.Deployment
		expectedPort int
	}

	newDeployment := func(portAnnotation string, portContainer int) *v1.Deployment {
		annotationMap := map[string]string{}
		if portAnnotation != "" {
			annotationMap["skupper.io/port"] = portAnnotation
		}

		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "dep1",
				Namespace:   "ns1",
				Annotations: annotationMap,
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Ports: []corev1.ContainerPort{{
								ContainerPort: int32(portContainer),
							}},
						}},
					},
				},
			},
		}
	}

	testTable := []test{
		{"no-annotation-container-port", newDeployment("", 8080), 8080},
		{"valid-annotation-container-port", newDeployment("8888", 8080), 8888},
		{"invalid-annotation-container-port", newDeployment("invalid", 8080), 0},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			assert.Assert(t, test.expectedPort == deducePort(test.deployment))
		})
	}
}

func TestDeducePortFromService(t *testing.T) {
	type test struct {
		name     string
		service  *corev1.Service
		expected int
	}

	// Helper used to prepare test table
	newService := func(ports ...int) *corev1.Service {
		// Only set ports, if at least one provided
		var svcPorts []corev1.ServicePort
		if len(ports) > 0 {
			for i, port := range ports {
				svcPorts = append(svcPorts, corev1.ServicePort{
					Name: fmt.Sprintf("port%d", i),
					Port: int32(port),
				})
			}
		}
		return &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: svcPorts,
			},
		}
	}

	testTable := []test{
		{"no-port", newService(), 0},
		{"one-port", newService(8080), 8080},
		{"two-ports", newService(8080, 8081), 8080},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, deducePortFromService(test.service), test.expected)
		})
	}
}

func TestDeduceTargetPortFromService(t *testing.T) {
	type test struct {
		name     string
		service  *corev1.Service
		expected int
	}

	// Helper used to prepare test table
	newService := func(ports ...int) *corev1.Service {
		// Only set ports, if at least one provided
		var svcPorts []corev1.ServicePort
		if len(ports) > 0 {
			for i, port := range ports {
				svcPorts = append(svcPorts, corev1.ServicePort{
					Name:       fmt.Sprintf("port%d", i),
					TargetPort: intstr.FromInt(port),
				})
			}
		}
		return &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: svcPorts,
			},
		}
	}

	testTable := []test{
		{"no-port", newService(), 0},
		{"one-port", newService(8080), 8080},
		{"two-ports", newService(8080, 8081), 8080},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, deduceTargetPortFromService(test.service), test.expected)
		})
	}
}
