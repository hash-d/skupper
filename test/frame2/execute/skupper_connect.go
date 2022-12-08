package execute

import (
	"context"
	"errors"
	"log"
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

	// Use this if the From namespace does not have a test runner
	// TODO
	RunnerBase *base.ClusterTestRunnerBase
}

func (sc SkupperConnect) Execute() error {
	log.Printf("execute.SkupperConnect")
	var err error

	ToCluster, err := sc.To.Satisfy()
	if err != nil {
		return err
	}
	fromCluster, err := sc.From.Satisfy()
	if err != nil {
		return err
	}

	log.Printf("connecting %v to %v", fromCluster.Namespace, ToCluster.Namespace)

	ctx := sc.Ctx

	var r base.ClusterTestRunnerBase
	if sc.RunnerBase != nil {
		r = *sc.RunnerBase
	} else {
		r := sc.From.Runner()
		if r == nil {
			return errors.New("SkupperConnect: empty runner on the From cluster")
		}
	}

	log.Printf("*")

	i := rand.Intn(1000)
	// TODO redo this file name: use domain names, but keep the random thing
	secretFile := "/tmp/" + r.Needs.NamespaceId + "_public_secret.yaml" + strconv.Itoa(i)
	err = ToCluster.VanClient.ConnectorTokenCreateFile(ctx, types.DefaultVanName, secretFile)
	if err != nil {
		return err
	}

	var connectorCreateOpts types.ConnectorCreateOptions = types.ConnectorCreateOptions{
		SkupperNamespace: fromCluster.Namespace,
		Name:             sc.Name,
		Cost:             sc.Cost,
	}
	_, err = fromCluster.VanClient.ConnectorCreateFromFile(ctx, secretFile, connectorCreateOpts)
	log.Printf("SkupperConnect done")
	return err

}
