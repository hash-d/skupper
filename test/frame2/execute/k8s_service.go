package execute

import (
	"fmt"
	"log"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Creates a Kubernetes service, with simplified configurations
type K8SServiceCreate struct {
	Namespace                *base.ClusterContextPromise
	Name                     string
	Annotations              map[string]string
	Labels                   map[string]string
	Selector                 map[string]string
	Ports                    []int32
	Type                     apiv1.ServiceType
	PublishNotReadyAddresses bool

	// Cluster IP; set this to "None" and Type to ClusterIP for a headless service
	// https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
	ClusterIP string

	AutoTeardown bool
	Wait         time.Duration
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
			Ports:                    ports,
			Selector:                 ks.Selector,
			Type:                     ks.Type,
			ClusterIP:                ks.ClusterIP,
			PublishNotReadyAddresses: ks.PublishNotReadyAddresses,
		},
	}

	// Creating the new service
	cluster, err := ks.Namespace.Satisfy()
	svc, err = cluster.VanClient.KubeClient.CoreV1().Services(cluster.Namespace).Create(svc)
	if err != nil {
		return err
	}
	if ks.Wait > 0 {
		log.Printf("Waiting up to %v until service %q exists", ks.Wait, ks.Name)
		_, err := frame2.Retry{
			Options: frame2.RetryOptions{
				KeepTrying: true,
				Timeout:    ks.Wait,
			},
			Fn: func() error {
				_, retryErr := cluster.VanClient.KubeClient.CoreV1().Services(cluster.Namespace).Get(ks.Name, metav1.GetOptions{})
				return retryErr
			},
		}.Run()

		return err
	}
	return nil
}

func (ks K8SServiceCreate) Teardown() frame2.Executor {

	if !ks.AutoTeardown {
		return nil

	}

	return K8SServiceDelete{
		Namespace: ks.Namespace,
		Name:      ks.Name,
	}
}

type K8SServiceDelete struct {
	Namespace *base.ClusterContextPromise
	Name      string
}

func (ksd K8SServiceDelete) Execute() error {
	cc, err := ksd.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy namespace promise: %w", err)
	}

	cc.VanClient.KubeClient.CoreV1().Services(cc.Namespace).Delete(ksd.Name, &metav1.DeleteOptions{})

	return nil
}

type K8SServiceAnnotate struct {
	Namespace   *base.ClusterContextPromise
	Name        string
	Annotations map[string]string
}

func (ksa K8SServiceAnnotate) Execute() error {
	cluster, err := ksa.Namespace.Satisfy()
	if err != nil {
		return err
	}
	// Retrieving service
	svc, err := cluster.VanClient.KubeClient.CoreV1().Services(cluster.VanClient.Namespace).Get(ksa.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if svc.Annotations == nil {
		svc.Annotations = map[string]string{}
	}

	for k, v := range ksa.Annotations {
		svc.Annotations[k] = v
	}
	_, err = cluster.VanClient.KubeClient.CoreV1().Services(cluster.Namespace).Update(svc)
	return err

}

type K8SServiceRemoveAnnotation struct {
	Namespace   *base.ClusterContextPromise
	Name        string
	Annotations []string
}

func (ksr K8SServiceRemoveAnnotation) Execute() error {
	cluster, err := ksr.Namespace.Satisfy()
	if err != nil {
		return err
	}
	// Retrieving service
	svc, err := cluster.VanClient.KubeClient.CoreV1().Services(cluster.VanClient.Namespace).Get(ksr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if svc.Annotations == nil {
		// Nothing to remove
		// TODO.  Perhaps a option to set an error if annotation not found to be removed
		return nil
	}

	for _, k := range ksr.Annotations {
		delete(svc.Annotations, k)
	}
	_, err = cluster.VanClient.KubeClient.CoreV1().Services(cluster.Namespace).Update(svc)
	return err

}

// Retrieve a K8S Service by name and namespace
type K8SServiceGet struct {
	Namespace *base.ClusterContextPromise
	Name      string

	// Return
	Service *apiv1.Service
}

func (kg *K8SServiceGet) Validate() error {
	cluster, err := kg.Namespace.Satisfy()
	kg.Service, err = cluster.VanClient.KubeClient.CoreV1().Services(cluster.Namespace).Get(kg.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}
