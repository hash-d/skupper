package execute

import (
	"github.com/skupperproject/skupper/test/utils/base"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Creates a Kubernetes service, with simplified configurations
type K8SServiceCreate struct {
	Namespace   *base.ClusterContextPromise
	Name        string
	Annotations map[string]string
	Labels      map[string]string
	Selector    map[string]string
	Ports       []int32
}

//func CreateService(cluster *client.VanClient, name string, annotations, labels, selector map[string]string, ports []apiv1.ServicePort) (*apiv1.Service, error) {

func (ks K8SServiceCreate) Execute() error {

	ports := []apiv1.ServicePort{}
	for _, port := range ks.Ports {
		ports = append(ports, apiv1.ServicePort{Port: port})
	}
	svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ks.Name,
			Labels:      ks.Labels,
			Annotations: ks.Annotations,
		},
		Spec: apiv1.ServiceSpec{
			Ports:    ports,
			Selector: ks.Selector,
			Type:     apiv1.ServiceTypeLoadBalancer,
		},
	}

	// Creating the new service

	//	cluster := ks.Namespace.VanClient
	cluster, err := ks.Namespace.Satisfy()
	svc, err = cluster.VanClient.KubeClient.CoreV1().Services(cluster.Namespace).Create(svc)
	if err != nil {
		return err
	}
	return nil
}
