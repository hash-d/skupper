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

	frame2.DefaultRunDealer
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
	Namespace     *base.ClusterContext
	EnableConsole bool

	frame2.DefaultRunDealer
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
				Modify: &CliSkupperInstall{
					Namespace:           sis.Namespace,
					EnableConsole:       sis.EnableConsole,
					EnableFlowCollector: sis.EnableConsole,
					//
				},
			},
		},
	}
	return phase.Run()
}

type CliSkupperInstall struct {
	Namespace           *base.ClusterContext
	Ctx                 context.Context
	MaxWait             time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait            bool
	SkipStatus          bool
	EnableConsole       bool
	EnableFlowCollector bool

	frame2.DefaultRunDealer
	SkupperVersionerDefault
}

// Interface execute.SkupperUpgradable; allow this to be used with Upgrade disruptors
func (s CliSkupperInstall) SkupperUpgradable() *base.ClusterContext {
	return s.Namespace
}

func (s CliSkupperInstall) Execute() error {
	versions := []string{"1.2", "1.3"}
	target := s.WhichSkupperVersion(versions)
	var action frame2.Executor
	switch target {
	case "1.3", "":
		action = &CliSkupperInstall_1_3{
			Namespace:           s.Namespace,
			Ctx:                 s.Ctx,
			MaxWait:             s.MaxWait,
			SkipWait:            s.SkipWait,
			EnableConsole:       s.EnableConsole,
			EnableFlowCollector: s.EnableFlowCollector,
		}
	case "1.2":
		action = &CliSkupperInstall_1_2{
			Namespace:     s.Namespace,
			Ctx:           s.Ctx,
			MaxWait:       s.MaxWait,
			SkipWait:      s.SkipWait,
			EnableConsole: s.EnableConsole,
		}
	default:
		panic("unnassigned version for CliSkupperInstall")
	}
	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: action,
			},
		},
	}

	return phase.Run()
}

type CliSkupperInstall_1_3 struct {
	Namespace           *base.ClusterContext
	Ctx                 context.Context
	MaxWait             time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait            bool
	SkipStatus          bool
	EnableConsole       bool
	EnableFlowCollector bool

	frame2.DefaultRunDealer
}

func (s CliSkupperInstall_1_3) Execute() error {

	args := []string{"init"}

	if s.EnableConsole {
		args = append(args, "--enable-console")
	}

	if s.EnableFlowCollector {
		args = append(args, "--enable-flow-collector")
	}

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
					Ctx:        s.Ctx,
				},
			},
		},
	}

	return phase.Run()
}

type CliSkupperInstall_1_2 struct {
	Namespace     *base.ClusterContext
	Ctx           context.Context
	MaxWait       time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait      bool
	SkipStatus    bool
	EnableConsole bool

	frame2.DefaultRunDealer
}

func (s CliSkupperInstall_1_2) Execute() error {

	args := []string{"init"}

	// On 1.3 the default changed from --enable-console=true to --enable-console=false.
	// For this reason, on 1.2 we need to always specify the console flag.
	args = append(args, fmt.Sprintf("--enable-console=%t", s.EnableConsole))

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

	frame2.DefaultRunDealer
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
					Args:           []string{"version"},
					ClusterContext: v.Namespace,
					Cmd: Cmd{
						ForceOutput: true,
					},
				},
				SkipWhen: v.SkipStatus,
			}, {
				Modify: &CliSkupper{
					Args:           []string{"status"},
					ClusterContext: v.Namespace,
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
