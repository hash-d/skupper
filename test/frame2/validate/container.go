package validate

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
	v1 "k8s.io/api/core/v1"
)

var (
	RouterSelector            = fmt.Sprintf("%s=%s", types.ComponentAnnotation, types.TransportComponentName)
	ConfigSyncSelector        = fmt.Sprintf("%s=%s", types.ComponentAnnotation, types.ConfigSyncContainerName)
	ServiceControllerSelector = fmt.Sprintf("%s=%s", types.ComponentAnnotation, types.ControllerComponentName)
)

type Container struct {
	Namespace        *base.ClusterContext
	PodSelector      string
	ContainerName    string // If empty, check all containers on selected pods
	ExpectNone       bool   // If true, it will be an error if any pods are found
	ExpectExactly    int    // if greater than 0, exactly this number of pods must be found
	CPULimit         string
	CPURequest       string
	MemoryLimit      string
	MemoryRequest    string
	CheckUnrequested bool // Normally an empty CPU/Memory Limit/Request is not checked
	RestartCount     int32
	RestartCheck     bool
	StatusCheck      bool

	Return *v1.Container
	frame2.Log
	frame2.DefaultRunDealer
}

func (c Container) Validate() error {
	return c.Run()
}

func (c Container) Run() error {
	log.Printf("Validating %+v", c)

	pods, err := kube.GetPods(c.PodSelector, c.Namespace.Namespace, c.Namespace.VanClient.KubeClient)
	if err != nil {
		return err
	}

	log.Printf("- Found %d pod(s)", len(pods))

	if c.ExpectNone {
		if len(pods) > 0 {
			return fmt.Errorf("expected no pods, found %d", len(pods))
		}
		return nil
	} else {
		if c.ExpectExactly > 0 {
			if len(pods) != c.ExpectExactly {
				return fmt.Errorf("expected exactly %d pods, found %d", c.ExpectExactly, len(pods))
			}
		} else {
			if len(pods) == 0 {
				return fmt.Errorf("expected at least one pod, found none")
			}
		}
	}

	// looping through pods
	for _, pod := range pods {
		log.Printf("- checking pod %q", pod.Name)

		var containerFound bool
		for _, container := range pod.Spec.Containers {
			if container.Name == c.ContainerName || c.ContainerName == "" {
				containerFound = true
				if c.ExpectExactly == 1 && len(pods) == 1 {
					c.Return = &container
				}

				log.Printf("- Checking container %v", container.Name)

				cpuRequest := container.Resources.Requests.Cpu().String()
				if c.CheckUnrequested || c.CPURequest != "" {
					log.Printf("- Validating CPURequest")
					if cpuRequest != c.CPURequest {
						return fmt.Errorf(
							"CPURequest %q different than expected %q",
							cpuRequest,
							c.CPURequest,
						)
					}
				}

				cpuLimit := container.Resources.Limits.Cpu().String()
				if c.CheckUnrequested || c.CPULimit != "" {
					log.Printf("- Validating CPULimit")
					if cpuLimit != c.CPULimit {
						return fmt.Errorf(
							"CPULimit %q different than expected %q",
							cpuLimit,
							c.CPULimit,
						)
					}
				}

				memoryRequest := container.Resources.Requests.Memory().String()
				if c.CheckUnrequested || c.MemoryRequest != "" {
					log.Printf("- Validating MemoryRequest")
					if memoryRequest != c.MemoryRequest {
						return fmt.Errorf(
							"MemoryRequest %q different than expected %q",
							memoryRequest,
							c.MemoryRequest,
						)
					}
				}

				memoryLimit := container.Resources.Limits.Memory().String()
				if c.CheckUnrequested || c.MemoryLimit != "" {
					log.Printf("- Validating MemoryLimit")
					if memoryLimit != c.MemoryLimit {
						return fmt.Errorf(
							"MemoryLimit %q different than expected %q",
							memoryLimit,
							c.MemoryLimit,
						)
					}
				}

			}
		}
		if !containerFound {
			return fmt.Errorf("container %q not found in pod %q", c.ContainerName, pod.Name)
		}

		for _, status := range pod.Status.ContainerStatuses {
			if c.RestartCheck && status.RestartCount != c.RestartCount && (c.ContainerName == "" || c.ContainerName == status.Name) {
				return fmt.Errorf("container %q has %d restarts, instead of the expected %d", status.Name, status.RestartCount, c.RestartCount)
			}

			if c.StatusCheck {
				// OCP 3.11 returns status.Started as nil, so the first check is required
				if status.Started != nil && !*status.Started {
					return fmt.Errorf("container %q (%v restarts) reports as not started", status.Name, status.RestartCount)
				}
				if !status.Ready {
					return fmt.Errorf("container %q (%v restarts) reports as not-ready", status.Name, status.RestartCount)
				}
			}
		}

	}
	return nil
}
