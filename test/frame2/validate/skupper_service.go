package validate

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperService struct {
	Namespace *base.ClusterContext
	Name      string

	Return *types.ServiceInterface
	frame2.Log
}

func (s SkupperService) Validate() error {

	list, err := s.Namespace.VanClient.ServiceInterfaceList(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	for _, item := range list {
		if item.Address == s.Name {
			s.Return = item
			return nil
		}
	}

	return fmt.Errorf("service %q not found in namespace %q", s.Name, s.Namespace.Namespace)
}
