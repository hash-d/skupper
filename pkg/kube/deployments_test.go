package kube_test

import (
	"fmt"
	"github.com/skupperproject/skupper/pkg/kube"
	"gotest.tools/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const (
	NS string = "ns1"
)

// TestGetContainerPort validates the behavior of GetContainerPort
// based on a test table with different inputs and expectedErr outcomes
func TestGetContainerPort(t *testing.T) {

	type test struct {
		name           string
		deployment     *v1.Deployment
		expectedResult int32
	}

	// helper function to compose test table
	newDeployment := func(name string, containers int, ports ...int) *v1.Deployment {
		containerPorts := []corev1.ContainerPort{}

		for idx, port := range ports {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				Name:          fmt.Sprintf("port%d", idx),
				ContainerPort: int32(port),
			})
		}

		depContainers := []corev1.Container{}
		for i := 1; i <= containers; i++ {
			depContainers = append(depContainers, corev1.Container{
				Name:  fmt.Sprintf("container%d", i),
				Ports: containerPorts,
			})
		}

		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: NS,
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: depContainers,
					},
				},
			},
		}
	}

	testTable := []test{
		{"no-container-no-port", newDeployment("dep0", 0), int32(0)},
		{"one-container-no-port", newDeployment("dep1", 1), int32(0)},
		{"one-container-one-port", newDeployment("dep1", 1, 8080), int32(8080)},
		{"one-container-multiple-ports", newDeployment("dep2", 1, 8080, 8081, 8082), int32(8080)},
		{"multiple-containers-multiple-ports", newDeployment("dep3", 3, 8080, 8081, 8082), int32(8080)},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			assert.Assert(t, test.expectedResult == kube.GetContainerPort(test.deployment))
		})
	}

}
