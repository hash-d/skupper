package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
)

// For a defaults alternative, check SkupperInstallSimple
type SkupperInstall struct {
	Namespace  *base.ClusterContext
	RouterSpec types.SiteConfigSpec
	Ctx        context.Context
	MaxWait    time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait   bool
	SkipStatus bool
	Runner     *frame2.Run
}

// Interface execute.SkupperUpgradable; allow this to be used with Upgrade disruptors
func (s SkupperInstall) SkupperUpgradable() *base.ClusterContext {
	return s.Namespace
}

// TODO: move this to a new SkupperInstallVAN or something; leave SkupperInstall as a
// SkupperOp that calls either that or CliSkupperInit
func (si SkupperInstall) Execute() error {

	ctx := si.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	wait := si.MaxWait
	if wait == 0 {
		wait = types.DefaultTimeoutDuration * 2
	}

	publicSiteConfig, err := si.Namespace.VanClient.SiteConfigCreate(ctx, si.RouterSpec)
	if err != nil {
		return fmt.Errorf("SkupperInstall failed to create SiteConfig: %w", err)
	}
	err = si.Namespace.VanClient.RouterCreate(ctx, *publicSiteConfig)
	if err != nil {
		return fmt.Errorf("SkupperInstall failed to create router: %w", err)
	}

	phase := frame2.Phase{
		Runner: si.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &ValidateSkupperAvailable{
					Namespace:  si.Namespace,
					MaxWait:    wait,
					SkipWait:   si.SkipStatus,
					SkipStatus: si.SkipStatus,
					Runner:     si.Runner,
					Ctx:        ctx,
				},
			},
		},
	}

	return phase.Run()

}

// A Skupper installation that uses some default configurations.
// It cannot be configured.  For a configurable version, use
// SkupperInstall, instead.
type SkupperInstallSimple struct {
	Namespace *base.ClusterContext
	Runner    *frame2.Run
}

func (sis SkupperInstallSimple) Execute() error {
	phase := frame2.Phase{
		Runner: sis.Runner,
		MainSteps: []frame2.Step{
			{
				//Modify: SkupperInstall{
				//	Runner:    sis.Runner,
				//	Namespace: sis.Namespace,
				//	RouterSpec: types.SiteConfigSpec{
				//		SkupperName:       "",
				//		RouterMode:        string(types.TransportModeInterior),
				//		EnableController:  true,
				//		EnableServiceSync: true,
				//		EnableConsole:     true,
				//		AuthMode:          types.ConsoleAuthModeInternal,
				//		User:              "admin",
				//		Password:          "admin",
				//		Ingress:           sis.Namespace.VanClient.GetIngressDefault(),
				//		Replicas:          1,
				//		Router:            constants.DefaultRouterOptions(nil),
				//	},
				//},
				Modify: CliSkupperInstall{
					Runner:    sis.Runner,
					Namespace: sis.Namespace,
					//
				},
			},
		},
	}
	return phase.Run()
}

type CliSkupperInstall struct {
	Namespace  *base.ClusterContext
	Ctx        context.Context
	MaxWait    time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait   bool
	SkipStatus bool
	Runner     *frame2.Run
}

// Interface execute.SkupperUpgradable; allow this to be used with Upgrade disruptors
func (s CliSkupperInstall) SkupperUpgradable() *base.ClusterContext {
	return s.Namespace
}

func (s CliSkupperInstall) Execute() error {

	args := []string{"init"}

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: s.Namespace,
				},
				Validator: &ValidateSkupperAvailable{
					Namespace:  s.Namespace,
					MaxWait:    s.MaxWait,
					SkipWait:   s.SkipStatus,
					SkipStatus: s.SkipStatus,
					Runner:     s.Runner,
					Ctx:        s.Ctx,
				},
			},
		},
	}

	return phase.Run()
}

type ValidateSkupperAvailable struct {
	Namespace  *base.ClusterContext
	Ctx        context.Context
	MaxWait    time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait   bool
	SkipStatus bool
	Runner     *frame2.Run

	frame2.Log
}

func (v ValidateSkupperAvailable) Validate() error {
	var waitCtx context.Context
	var cancel context.CancelFunc

	wait := v.MaxWait
	if wait == 0 {
		wait = 2 * time.Minute
	}

	if !v.SkipWait {
		waitCtx, cancel = context.WithTimeout(v.Runner.OrDefaultContext(v.Ctx), wait)
		defer cancel()
	}

	phase := frame2.Phase{
		Runner: v.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Check that the router and service controller containers are reporting as ready",
				Validators: []frame2.Validator{
					&validate.Container{
						Namespace:   v.Namespace,
						PodSelector: validate.RouterSelector,
						StatusCheck: true,
					},
					&validate.Container{
						Namespace:   v.Namespace,
						PodSelector: validate.ServiceControllerSelector,
						StatusCheck: true,
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Ctx:        waitCtx,
					Ensure:     5, // The containers may briefly report ready before crashing
					KeepTrying: true,
				},
				SkipWhen: v.SkipWait,
			}, {
				Modify: &CliSkupper{
					Args:      []string{"version"},
					Namespace: v.Namespace.Namespace,
					Cmd: Cmd{
						ForceOutput: true,
					},
				},
				SkipWhen: v.SkipStatus,
			}, {
				Modify: &CliSkupper{
					Args:      []string{"status"},
					Namespace: v.Namespace.Namespace,
					Cmd: Cmd{
						ForceOutput: true,
					},
				},
				SkipWhen: v.SkipStatus,
			},
		},
	}
	return phase.Run()
}
