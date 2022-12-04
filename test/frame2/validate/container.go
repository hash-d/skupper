package validate

import (
	"fmt"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/frame2"
)

var (
	RouterSelector     = fmt.Sprintf("%s=%s", types.ComponentAnnotation, types.TransportComponentName)
	ConfigSyncSelector = fmt.Sprintf("%s=%s", types.ComponentAnnotation, types.ConfigSyncContainerName)
)

type Container struct {
	frame2.Validate
	PodSelector      string
	ContainerName    string // If empty, check all containers on selected pods
	ExpectNone       bool   // If true, it will be an error if any pods are found
	ExpectExactly    int    // if greater than 0, exactly this number of pods must be found
	CPULimit         string
	CPURequest       string
	MemoryLimit      string
	MemoryRequest    string
	CheckUnrequested bool // Normally an empty CPU/Memory Limit/Request is not checked
}

func (c Container) Run() error {
	c.Logf("Validating %+v", c)

	cluster, err := c.Namespace.Satisfy()
	if err != nil {
		return err
	}

	// retrieving the router pods
	pods, err := kube.GetPods(c.PodSelector, cluster.Namespace, cluster.VanClient.KubeClient)
	if err != nil {
		return err
	}

	c.Logf("- Found %d pod(s)", len(pods))

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
		c.Logf("- checking pod %q", pod.Name)
		var containerFound bool
		for _, container := range pod.Spec.Containers {
			if container.Name == c.ContainerName || c.ContainerName == "" {
				containerFound = true
				c.Logf("- Checking container %v", container.Name)

				cpuRequest := container.Resources.Requests.Cpu().String()
				if c.CheckUnrequested || c.CPURequest != "" {
					c.Logf("- Validating CPURequest")
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
					c.Logf("- Validating CPULimit")
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
					c.Logf("- Validating MemoryRequest")
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
					c.Logf("- Validating MemoryLimit")
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
	}
	return nil
}
