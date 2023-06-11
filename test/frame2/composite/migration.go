package composite

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Migrate an application and Skupper out of a
// cluster context and into another
//
// - Deploy the application
// - Install Skupper
// - Create skupper links (to/from target cctx)
// - Remove Skupper from old namespace
// - Remove application from old namespace
//
// Note the application deployment can be done as the very first
// step or after the link step (for the situations, for example,
// where the application depends on other services on the VAN)
type Migrate struct {
	From                *base.ClusterContext
	To                  *base.ClusterContext
	DeploySteps         []frame2.Step
	UndeploySteps       []frame2.Step
	LinkTo              []*base.ClusterContext
	LinkFrom            []*base.ClusterContext
	UnlinkFrom          []*base.ClusterContext // Avoids dangling link configuration
	DeployBeforeSkupper bool
	AssertFromEmpty     bool

	frame2.DefaultRunDealer

	// Application validation?
	// TODO: change (Un)DeploySteps from frame2.Step to new frame2.TargetSetter
	//       TargetSetter is a Step that has a SetTarget (*base.ClusterContext)
	//       function, which sets its target cctx
}

func (m *Migrate) Execute() error {

	deployPhase := frame2.Phase{
		Runner:    m.Runner,
		MainSteps: m.DeploySteps,
	}
	if m.DeployBeforeSkupper {
		deployPhase.Run()
	}

	skupperInstallPhase := frame2.Phase{
		Runner: m.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: fmt.Sprintf("Install Skupper on new namespace %q", m.To.Namespace),
				Modify: &execute.SkupperInstallSimple{
					Namespace: m.To,
				},
			},
		},
	}
	skupperInstallPhase.Run()

	type linkStruct struct {
		from *base.ClusterContext
		to   *base.ClusterContext
	}

	links := []linkStruct{}

	for _, i := range m.LinkTo {
		links = append(links, linkStruct{m.To, i})
	}
	for _, i := range m.LinkFrom {
		links = append(links, linkStruct{i, m.To})
	}

	var linkSteps []frame2.Step

	for _, l := range links {
		linkSteps = append(linkSteps, frame2.Step{
			Doc: fmt.Sprintf("connecting %v to %v", l.from.Namespace, l.to.Namespace),
			Modify: execute.SkupperConnect{
				Name: fmt.Sprintf("%v-to-%v", l.from.Namespace, l.to.Namespace),
				From: l.from,
				To:   l.to,
			},
		})
	}
	linkPhase := frame2.Phase{
		Runner:    m.Runner,
		MainSteps: linkSteps,
	}
	linkPhase.Run()

	if !m.DeployBeforeSkupper {
		deployPhase.Run()
	}

	var unlinkSteps []frame2.Step
	for _, l := range m.UnlinkFrom {
		unlinkSteps = append(unlinkSteps, frame2.Step{
			Doc: fmt.Sprintf("removing link from %v to %v", l.Namespace, m.From.Namespace),
			Modify: execute.SkupperUnLink{
				Name:   fmt.Sprintf("%v-to-%v", l.Namespace, m.From.Namespace),
				From:   l,
				To:     m.From,
				Runner: m.Runner,
			},
		})
	}

	unlinkPhase := frame2.Phase{
		Runner:    m.Runner,
		MainSteps: unlinkSteps,
	}
	unlinkPhase.Run()

	removalPhase := frame2.Phase{
		Runner: m.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "remove skupper from the old namespace",
				Modify: &execute.SkupperDelete{
					Namespace: m.From,
				},
			},
		},
	}
	removalPhase.MainSteps = append(removalPhase.MainSteps, m.UndeploySteps...)

	// Add step K8SCheckNamespaceIsEmpty.  Check for deployments,
	// secrets, configmaps, services, etc
	return removalPhase.Run()
}
