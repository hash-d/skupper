package execute

import (
	"context"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperUnLink struct {
	Name   string
	From   *base.ClusterContext
	To     *base.ClusterContext
	Ctx    context.Context
	Runner *frame2.Run
	frame2.Log
}

func (s SkupperUnLink) Execute() error {

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					ClusterContext: s.From,
					Args:           []string{"link", "delete", s.Name},
				},
			},
		},
	}
	return phase.Run()
}
