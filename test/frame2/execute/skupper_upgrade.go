package execute

import (
	"context"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperUpgrade struct {
	Runner       *frame2.Run
	Namespace    *base.ClusterContext
	ForceRestart bool
	SkipVersion  bool

	Wait time.Duration
	Ctx  context.Context
}

func (s SkupperUpgrade) Execute() error {

	args := []string{"update"}

	if s.ForceRestart {
		args = append(args, "--force-restart")
	}

	ctx := s.Runner.OrDefaultContext(s.Ctx)
	var cancel context.CancelFunc
	var validators []frame2.Validator
	if s.Wait != 0 {
		ctx, cancel = context.WithTimeout(s.Runner.OrDefaultContext(ctx), s.Wait)
		defer cancel()

		validators = []frame2.Validator{
			&validate.Container{
				Namespace:   s.Namespace,
				PodSelector: validate.RouterSelector,
				StatusCheck: true,
			},
			&validate.Container{
				Namespace:   s.Namespace,
				PodSelector: validate.ServiceControllerSelector,
				StatusCheck: true,
			},
		}
	}

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					ClusterContext: s.Namespace,
					Args:           args,
					Cmd: Cmd{
						Ctx: ctx,
					},
				},
				Validators: validators,
				ValidatorRetry: frame2.RetryOptions{
					Allow:      60,
					Ignore:     10,
					Ensure:     5,
					KeepTrying: true,
					Ctx:        ctx,
				},
			}, {
				Modify: &CliSkupper{
					ClusterContext: s.Namespace,
					Args:           []string{"version"},
				},
				SkipWhen: s.SkipVersion,
			},
		},
	}
	return phase.Run()
}
