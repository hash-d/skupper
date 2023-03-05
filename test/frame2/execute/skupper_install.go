package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/constants"
)

type SkupperInstall struct {
	Namespace  *base.ClusterContextPromise
	RouterSpec types.SiteConfigSpec
	Ctx        context.Context
	MaxWait    time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait   bool
	Runner     frame2.Run
}

func (si SkupperInstall) Execute() error {

	ctx := si.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	wait := si.MaxWait
	if wait == 0 {
		wait = types.DefaultTimeoutDuration * 2
	}

	cluster, err := si.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("SkupperInstall failed to satisfy namespace promise: %w", err)
	}
	publicSiteConfig, err := cluster.VanClient.SiteConfigCreate(ctx, si.RouterSpec)
	if err != nil {
		return fmt.Errorf("SkupperInstall failed to create SiteConfig: %w", err)
	}
	err = cluster.VanClient.RouterCreate(ctx, *publicSiteConfig)
	if err != nil {
		return fmt.Errorf("SkupperInstall failed to create router: %w", err)
	}

	if !si.SkipWait {
		waitCtx, cancel := context.WithTimeout(ctx, wait)
		defer cancel()

		phase := frame2.Phase{
			Runner: &si.Runner,
			MainSteps: []frame2.Step{{
				Doc: "Check that the router and service controller containers are reporting as ready",
				Validators: []frame2.Validator{
					validate.Container{
						Namespace:   si.Namespace,
						PodSelector: validate.RouterSelector,
						StatusCheck: true,
					},
					validate.Container{
						Namespace:   si.Namespace,
						PodSelector: validate.ServiceControllerSelector,
						StatusCheck: true,
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Ctx:        waitCtx,
					Ensure:     5, // The containers may briefly report ready before crashing
					KeepTrying: true,
				},
			}},
		}
		return phase.Run()
	}

	return nil
}

// A Skupper installation that uses some default configurations.
// It cannot be configured.  For a configurable version, use
// SkupperInstall, instead.
type SkupperInstallSimple struct {
	Namespace *base.ClusterContextPromise
}

func (sis SkupperInstallSimple) Execute() error {
	cluster, err := sis.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("SkupperInstallSimple failed to get cluster: %w", err)
	}
	si := SkupperInstall{
		Namespace: sis.Namespace,
		RouterSpec: types.SiteConfigSpec{
			SkupperName:       "",
			RouterMode:        string(types.TransportModeInterior),
			EnableController:  true,
			EnableServiceSync: true,
			EnableConsole:     true,
			AuthMode:          types.ConsoleAuthModeInternal,
			User:              "admin",
			Password:          "admin",
			Ingress:           cluster.VanClient.GetIngressDefault(),
			Replicas:          1,
			Router:            constants.DefaultRouterOptions(nil),
		},
	}
	return si.Execute()
}