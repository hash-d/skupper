package execute

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperConnect struct {
	Name string
	Cost int32
	From *base.ClusterContextPromise
	To   *base.ClusterContextPromise
	Ctx  context.Context
}

func (sc SkupperConnect) Execute() error {
	var err error

	ToCluster, err := sc.To.Satisfy()
	if err != nil {
		return err
	}
	fromCluster, err := sc.From.Satisfy()
	if err != nil {
		return err
	}

	ctx := sc.Ctx
	r := sc.From.Runner()
	if r == nil {
		return fmt.Errorf("SkupperConnect: empty runner on the From cluster")
	}

	/*

		routerCreateSpecFrom := types.SiteConfigSpec{
			SkupperName:       "",
			RouterMode:        string(types.TransportModeEdge),
			EnableController:  true,
			EnableServiceSync: true,
			EnableConsole:     true,
			AuthMode:          types.ConsoleAuthModeUnsecured,
			User:              "admin",
			Password:          "admin",
			Ingress:           ToCluster.VanClient.GetIngressDefault(),
			Replicas:          1,
			Router:            constants.DefaultRouterOptions(nil),
		}

		testContext, cancel := context.WithTimeout(ctx, types.DefaultTimeoutDuration*2)
		defer cancel()

		publicSiteConfig, err := ToCluster.VanClient.SiteConfigCreate(context.Background(), routerCreateSpecTo)
		if err != nil {
			return err
		}

		err = ToCluster.VanClient.RouterCreate(testContext, *publicSiteConfig)
		if err != nil {
			return err
		}
	*/

	i := rand.Intn(1000)
	// TODO redo this file name: use domain names, but keep the random thing
	secretFile := "/tmp/" + r.Needs.NamespaceId + "_public_secret.yaml" + strconv.Itoa(i)
	err = ToCluster.VanClient.ConnectorTokenCreateFile(ctx, types.DefaultVanName, secretFile)
	if err != nil {
		return err
	}

	/*
		// Configure private cluster.
		routerCreateSpecFrom.SkupperNamespace = fromCluster.Namespace
		privateSiteConfig, err := fromCluster.VanClient.SiteConfigCreate(context.Background(), routerCreateSpecFrom)

		err = fromCluster.VanClient.RouterCreate(testContext, *privateSiteConfig)
		if err != nil {
			return err
		}
	*/

	var connectorCreateOpts types.ConnectorCreateOptions = types.ConnectorCreateOptions{
		SkupperNamespace: fromCluster.Namespace,
		Name:             sc.Name,
		Cost:             sc.Cost,
	}
	_, err = fromCluster.VanClient.ConnectorCreateFromFile(ctx, secretFile, connectorCreateOpts)
	return err

}
