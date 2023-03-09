package execute

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Executes a fully specified K8S Statefulset
type K8SStatefulSet struct {
	Namespace    *base.ClusterContextPromise
	StatefulSet  *apps.StatefulSet
	AutoTeardown bool
	Ctx          context.Context

	Result *apps.StatefulSet
}

func (k *K8SStatefulSet) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)
	cc, err := k.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}

	k.Result, err = cc.VanClient.KubeClient.AppsV1().StatefulSets(cc.Namespace).Create(ctx, k.StatefulSet, metav1.CreateOptions{})
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

	Ctx context.Context
}

func (k *K8SStatefulSetRemove) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)
	cc, err := k.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("Failed to satisfy ClusterContextPromise: %w", err)
	}

	err = cc.VanClient.KubeClient.AppsV1().StatefulSets(cc.Namespace).Delete(ctx, k.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove statefulset %q: %w", k.Name, err)
	}

	return nil
}
