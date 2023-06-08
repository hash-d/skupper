package execute

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperDelete struct {
	Namespace *base.ClusterContext

	Context context.Context
}

// TODO: remove autodebug
func (s *SkupperDelete) Execute() error {

	ctx := s.Context
	if s.Context == nil {
		ctx = context.Background()
	}

	err := s.Namespace.VanClient.SiteConfigRemove(ctx)
	if err != nil {
		return fmt.Errorf("SkupperDelete failed to remove SiteConfig: %w", err)
	}

	// This is done automatically when site config is removed.
	// See https://github.com/skupperproject/skupper/issues/765
	// 	err = s.Namespace.VanClient.RouterRemove(ctx)
	// 	if err != nil {
	// 		return fmt.Errorf("SkupperDelete failed to remove Router: %w", err)
	// 	}

	return nil
}
