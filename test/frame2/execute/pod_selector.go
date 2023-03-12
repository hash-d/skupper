package execute

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/utils/base"
	v1 "k8s.io/api/core/v1"
)

type PodSelector struct {
	Namespace     base.ClusterContext
	Selector      string
	ExpectNone    bool // If true, it will be an error if any pods are found
	ExpectExactly int  // if greater than 0, exactly this number of pods must be found

	// Return value
	Pods []v1.Pod
}

func (p PodSelector) Execute() error {

	pods, err := kube.GetPods(p.Selector, p.Namespace.Namespace, p.Namespace.VanClient.KubeClient)
	if err != nil {
		return err
	}

	log.Printf("- Found %d pod(s)", len(pods))

	if p.ExpectNone {
		if len(pods) > 0 {
			return fmt.Errorf("expected no pods, found %d", len(pods))
		}
		return nil
	} else {
		if p.ExpectExactly > 0 {
			if len(pods) != p.ExpectExactly {
				return fmt.Errorf("expected exactly %d pods, found %d", p.ExpectExactly, len(pods))
			}
		} else {
			if len(pods) == 0 {
				return fmt.Errorf("expected at least one pod, found none")
			}
		}
	}

	p.Pods = append(p.Pods, pods...)

	return nil
}
