package execute

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/constants"
)

type SkupperInstall struct {
	Namespace  *base.ClusterContextPromise
	RouterSpec types.SiteConfigSpec
}

func (si SkupperInstall) Execute() error {
	cluster, err := si.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("SkupperInstall failed: %w", err)
	}
	testContext, cancel := context.WithTimeout(context.Background(), types.DefaultTimeoutDuration*2)
	defer cancel()
	publicSiteConfig, err := cluster.VanClient.SiteConfigCreate(context.Background(), si.RouterSpec)
	if err != nil {
		return fmt.Errorf("SkupperInstall failed: %w", err)
	}
	err = cluster.VanClient.RouterCreate(testContext, *publicSiteConfig)
	if err != nil {
		return err
	}
	return nil
}

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
