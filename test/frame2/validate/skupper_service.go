package validate

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperService struct {
	Namespace *base.ClusterContextPromise
	Name      string

	Return *types.ServiceInterface
}

func (ss SkupperService) Validate() error {

	namespace, err := ss.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("failed to satisfy namespace: %w", err)
	}

	list, err := namespace.VanClient.ServiceInterfaceList(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	for _, item := range list {
		if item.Address == ss.Name {
			ss.Return = item
			return nil
		}
	}

	return fmt.Errorf("service %q not found in namespace %q", ss.Name, namespace.Namespace)
}
