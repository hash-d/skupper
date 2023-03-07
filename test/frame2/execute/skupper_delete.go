package execute

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperDelete struct {
	Namespace *base.ClusterContextPromise

	Context context.Context
}

// TODO: remove autodebug
func (s *SkupperDelete) Execute() error {

	ctx := s.Context
	if s.Context == nil {
		ctx = context.Background()
	}

	cluster, err := s.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("SkupperDelete failed to satisfy namespace promise: %w", err)
	}

	err = cluster.VanClient.SiteConfigRemove(ctx)
	if err != nil {
		return fmt.Errorf("SkupperDelete failed to remove SiteConfig: %w", err)
	}

	err = cluster.VanClient.RouterRemove(ctx)
	if err != nil {
		return fmt.Errorf("SkupperDelete failed to remove Router: %w", err)
	}

	return nil
}
