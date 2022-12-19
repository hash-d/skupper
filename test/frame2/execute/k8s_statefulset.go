package execute

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Executes a fully specified K8S Statefulset
//
type K8SStatefulSet struct {
	Namespace    *base.ClusterContextPromise
	StatefulSet  *apps.StatefulSet
	AutoTeardown bool

	Result *apps.StatefulSet
}

func (k *K8SStatefulSet) Execute() error {
	cc, err := k.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}

	k.Result, err = cc.VanClient.KubeClient.AppsV1().StatefulSets(cc.Namespace).Create(k.StatefulSet)
	if err != nil {
		return fmt.Errorf("Failed to create statefulset %q: %w", k.StatefulSet.Name, err)
	}

	return nil
}

func (k *K8SStatefulSet) Teardown() frame2.Executor {
	if !k.AutoTeardown || k.StatefulSet == nil {
		return nil
	}

	return &K8SStatefulSetRemove{
		Namespace: k.Namespace,
		Name:      k.StatefulSet.Name,
	}

}

type K8SStatefulSetRemove struct {
	Namespace *base.ClusterContextPromise
	Name      string
}

func (k *K8SStatefulSetRemove) Execute() error {
	cc, err := k.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}

	err = cc.VanClient.KubeClient.AppsV1().StatefulSets(cc.Namespace).Delete(k.Name, &meta.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove statefulset %q: %w", k.Name, err)
	}

	return nil
}
