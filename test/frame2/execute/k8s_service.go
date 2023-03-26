package execute

import (
	"context"
	"log"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Creates a Kubernetes service, with simplified configurations
type K8SServiceCreate struct {
	Namespace                *base.ClusterContext
	Name                     string
	Annotations              map[string]string
	Labels                   map[string]string
	Selector                 map[string]string
	Ports                    []int32
	Type                     apiv1.ServiceType
	PublishNotReadyAddresses bool
	Ctx                      context.Context

	// Cluster IP; set this to "None" and Type to ClusterIP for a headless service
	// https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
	ClusterIP string

	AutoTeardown bool
	Wait         time.Duration
}

//func CreateService(cluster *client.VanClient, name string, annotations, labels, selector map[string]string, ports []apiv1.ServicePort) (*apiv1.Service, error) {

func (ks K8SServiceCreate) Execute() error {
	ctx := frame2.ContextOrDefault(ks.Ctx)

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
	svc, err := ks.Namespace.VanClient.KubeClient.CoreV1().Services(ks.Namespace.Namespace).Create(ctx, svc, v1.CreateOptions{})
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
				_, retryErr := ks.Namespace.VanClient.KubeClient.CoreV1().Services(ks.Namespace.Namespace).Get(ctx, ks.Name, metav1.GetOptions{})
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
	Namespace *base.ClusterContext
	Name      string

	Ctx context.Context
}

func (ksd K8SServiceDelete) Execute() error {
	ctx := frame2.ContextOrDefault(ksd.Ctx)

	ksd.Namespace.VanClient.KubeClient.CoreV1().Services(ksd.Namespace.Namespace).Delete(ctx, ksd.Name, metav1.DeleteOptions{})

	return nil
}

type K8SServiceAnnotate struct {
	Namespace   *base.ClusterContext
	Name        string
	Annotations map[string]string

	Ctx context.Context
}

func (ksa K8SServiceAnnotate) Execute() error {
	ctx := frame2.ContextOrDefault(ksa.Ctx)
	// Retrieving service
	svc, err := ksa.Namespace.VanClient.KubeClient.CoreV1().Services(ksa.Namespace.VanClient.Namespace).Get(ctx, ksa.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if svc.Annotations == nil {
		svc.Annotations = map[string]string{}
	}

	for k, v := range ksa.Annotations {
		svc.Annotations[k] = v
	}
	_, err = ksa.Namespace.VanClient.KubeClient.CoreV1().Services(ksa.Namespace.Namespace).Update(ctx, svc, v1.UpdateOptions{})
	return err

}

type K8SServiceRemoveAnnotation struct {
	Namespace   *base.ClusterContext
	Name        string
	Annotations []string

	Ctx context.Context
}

func (ksr K8SServiceRemoveAnnotation) Execute() error {
	ctx := frame2.ContextOrDefault(ksr.Ctx)
	// Retrieving service
	svc, err := ksr.Namespace.VanClient.KubeClient.CoreV1().Services(ksr.Namespace.VanClient.Namespace).Get(ctx, ksr.Name, metav1.GetOptions{})
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
	_, err = ksr.Namespace.VanClient.KubeClient.CoreV1().Services(ksr.Namespace.Namespace).Update(ctx, svc, v1.UpdateOptions{})
	return err

}

// Retrieve a K8S Service by name and namespace
type K8SServiceGet struct {
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	frame2.Log

	// Return
	Service *apiv1.Service
}

func (kg *K8SServiceGet) Validate() error {
	ctx := frame2.ContextOrDefault(kg.Ctx)
	var err error
	kg.Service, err = kg.Namespace.VanClient.KubeClient.CoreV1().Services(kg.Namespace.Namespace).Get(ctx, kg.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}
