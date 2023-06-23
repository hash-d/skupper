package execute

import (
	"context"
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type PostgresPing struct {
	Namespace *base.ClusterContext
	Podname   string
	Labels    map[string]string
	Container string

	DbName   string
	DbHost   string
	DbPort   string // default is 5432
	Username string

	Ctx context.Context

	frame2.Log
	frame2.DefaultRunDealer
}

func (p *PostgresPing) Validate() error {

	port := p.DbPort
	if port == "" {
		port = "5432"
	}

	command := []string{
		"pg_isready",
		fmt.Sprintf("--dbname=%v", p.DbName),
		fmt.Sprintf("--host=%v", p.DbHost),
		fmt.Sprintf("--port=%v", port),
	}
	if p.Username != "" {
		command = append(command, fmt.Sprintf("--username=%v", p.Username))
	}

	phase := frame2.Phase{
		Runner: p.Runner,
		Log:    p.Log,
		MainSteps: []frame2.Step{
			{
				Validator: &K8SPodExecute{
					Pod: &K8SPodGet{
						Namespace: p.Namespace,
						Labels:    p.Labels,
					},
					Container: p.Container,
					Command:   command,
					Ctx:       p.Ctx,
					Log:       p.Log,
				},
			},
		},
	}

	return phase.Run()

}
