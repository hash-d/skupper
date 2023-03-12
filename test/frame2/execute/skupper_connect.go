package execute

import (
	"context"
	"log"
	"math/rand"
	"strconv"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Connects two Skupper instances installed in different namespaces or clusters
//
// In practice, it does two steps: create the token, then use it to create a link
// on the other namespace
type SkupperConnect struct {
	Name string
	Cost int32
	From *base.ClusterContext
	To   *base.ClusterContext
	Ctx  context.Context
}

func (sc SkupperConnect) Execute() error {
	ctx := frame2.ContextOrDefault(sc.Ctx)

	log.Printf("execute.SkupperConnect")
	var err error

	log.Printf("connecting %v to %v", sc.From.Namespace, sc.To.Namespace)

	i := rand.Intn(1000)
	secretFile := "/tmp/" + sc.To.Namespace + "_secret.yaml." + strconv.Itoa(i)
	err = sc.To.VanClient.ConnectorTokenCreateFile(ctx, types.DefaultVanName, secretFile)
	if err != nil {
		return err
	}

	var connectorCreateOpts types.ConnectorCreateOptions = types.ConnectorCreateOptions{
		SkupperNamespace: sc.From.Namespace,
		Name:             sc.Name,
		Cost:             sc.Cost,
	}
	_, err = sc.From.VanClient.ConnectorCreateFromFile(ctx, secretFile, connectorCreateOpts)
	log.Printf("SkupperConnect done")
	return err

}
