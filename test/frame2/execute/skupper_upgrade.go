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

	// TODO
	// If true, skips checking the images against the manifest.  If
	// false and no manifest available, panic
	SkipManifest bool

	// Location of the manifest file to be used on the manifest/image
	// tag check.  If empty, check: (?)
	// - Current dir (ie, etc package dir)
	// - Source root dir
	// TODO
	// - Same directory as the skupper binary
	ManifestFile string

	// TODO: SkupperBinary (for multi-step upgrades)
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
	err := phase.Run()
	if err != nil {
		return err
	}
	if s.SkipManifest {
		return nil
	}

	skupperInfo := validate.SkupperInfo{
		Namespace: s.Namespace,
		Ctx:       s.Ctx,
	}
	getInfoPhase := frame2.Phase{
		Runner: s.Runner,
		Doc:    "Get the newly-upgrade Skupper info",
		MainSteps: []frame2.Step{
			{
				Validator: &skupperInfo,
			},
		},
	}
	err = getInfoPhase.Run()
	if err != nil {
		return err
	}

	checkManifestPhase := frame2.Phase{
		Runner: s.Runner,
		Doc:    "Compare Skupper images to the manifest.json",
		MainSteps: []frame2.Step{
			{
				Validator: &validate.SkupperManifest{
					Expected: skupperInfo.Result.Images,
				},
			},
		},
	}
	return checkManifestPhase.Run()
}
