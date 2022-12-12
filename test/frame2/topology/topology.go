package topology

import (
	"errors"
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

type ClusterType int

const (
	Public ClusterType = iota
	Private
	DMZ
)

type TopologyItem struct {
	Type        ClusterType
	Connections []*TopologyItem
	//SkipSkupperDeploy bool // TODO
}

// clients should keep a reference to a TopologyMap to
// get their output
type TopologyMap struct {
	Name           string
	TestRunnerBase *base.ClusterTestRunnerBase

	// Input
	Map []*TopologyItem

	// Output
	Private []*base.ClusterContext
	Public  []*base.ClusterContext

	GeneratedMap map[*TopologyItem]*base.ClusterContext
}

// Validate: check for duplicates, disconnected items, etc (but allow to skip validation)
func (tm *TopologyMap) Execute() error {
	if tm.Name == "" {
		return fmt.Errorf("TopologyMap configurarion error: no name provided")
	}
	if len(tm.Map) == 0 {
		return fmt.Errorf("TopologyMap configuration error: no topology provided")
	}
	err := TopologyValidator{}.Execute()
	if err != nil {
		return err
	}

	countPrivate := 0
	countPublic := 0
	for _, item := range tm.Map {
		switch item.Type {
		case Public:
			countPublic++
		case Private:
			countPrivate++
		default:
			return fmt.Errorf("TopologyMap: only Public and Private implemented")
		}
	}

	needs := base.ClusterNeeds{
		NamespaceId:     tm.Name,
		PublicClusters:  countPublic,
		PrivateClusters: countPrivate,
	}

	err = tm.TestRunnerBase.Validate(needs)
	if err != nil {
		return fmt.Errorf("TopologyMap: failed validating needs: %w", err)
	}

	_, err = tm.TestRunnerBase.Build(needs, nil)
	if err != nil {
		return fmt.Errorf("TopologyMap: failed building the contexts: %w", err)
	}

	tm.GeneratedMap = map[*TopologyItem]*base.ClusterContext{}

	countPrivate = 0
	countPublic = 0
	for _, item := range tm.Map {

		switch item.Type {
		case Public:
			countPublic++
			newContext, err := tm.TestRunnerBase.GetPublicContext(countPublic)
			if err != nil {
				return fmt.Errorf("TopologyMap failed to get public context %d: %w", countPublic, err)
			}
			tm.Public = append(tm.Public, newContext)
			tm.GeneratedMap[item] = newContext
		case Private:
			countPrivate++
			newContext, err := tm.TestRunnerBase.GetPrivateContext(countPrivate)
			if err != nil {
				return fmt.Errorf("TopologyMap failed to get public context %d: %w", countPrivate, err)
			}
			tm.Private = append(tm.Private, newContext)
			tm.GeneratedMap[item] = newContext
		default:
			return errors.New("TopologyMap: Only Public and Private implemented")
		}
	}

	return nil
}

type TopologyValidator struct {
	TopologyMap
}

func (tv TopologyValidator) Execute() error {
	return nil
}

// Creates a full topology: clusters, namespaces,
// skupper installations and the links between them
//
// This ties together TopologyMap, TopologyConnect
// and other items
type Topology struct {
	Runner       *frame2.Run
	TopologyMap  *TopologyMap
	AutoTearDown bool

	teardowns []frame2.Executor
}

func (t *Topology) Execute() error {
	steps := frame2.Phase{
		Runner: t.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: t.TopologyMap,
			},
		},
	}
	steps.Execute()

	log.Printf("Creating namespaces and installing Skupper")

	for _, context := range append(t.TopologyMap.Private, t.TopologyMap.Public...) {
		cc := context.GetPromise()

		createAndInstall := frame2.Phase{
			Runner: t.Runner,
			Setup: []frame2.Step{
				{
					Modify: execute.TestRunnerCreateNamespace{
						Namespace:    *cc,
						AutoTearDown: t.AutoTearDown,
					},
				}, {
					Modify: execute.SkupperInstallSimple{
						Namespace: cc,
					},
				},
			},
		}
		createAndInstall.Execute()

	}

	connectSteps := frame2.Phase{
		Runner: t.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: TopologyConnect{
					TopologyMap: *t.TopologyMap,
				},
			},
		},
	}
	connectSteps.Execute()

	return nil
}

// TODO Perhaps change the frame2.TearDowner interface to return a []frame2.Executor, instead, so a single
// call may return several, and have them run by the Runner?
func (t Topology) TearDown() frame2.Executor {
	return execute.Function{
		Fn: func() error {
			var ret error
			for _, td := range t.teardowns {
				err := td.Execute()
				if err != nil {
					log.Printf("topology teardown failed: %v", err)
					ret = fmt.Errorf("at least one step of topology teardown failed.  Last error: %w", err)
				}

			}
			return ret
		},
	}
}

type TopologyConnect struct {
	TopologyMap TopologyMap
}

// Assumes that the namespaces are already created, and Skupper installed on all
// namespaces that will create or receive links
func (tc TopologyConnect) Execute() error {

	for from, ctx := range tc.TopologyMap.GeneratedMap {
		for _, to := range from.Connections {
			pivot := tc.TopologyMap.GeneratedMap[to]
			connName := fmt.Sprintf("%v-to-%v", ctx.Namespace, pivot.Namespace)
			log.Printf("TopologyConnect creating connection from %v", connName)
			err := execute.SkupperConnect{
				Name:       connName,
				From:       ctx.GetPromise(),
				To:         pivot.GetPromise(),
				RunnerBase: tc.TopologyMap.TestRunnerBase,
			}.Execute()
			if err != nil {
				return fmt.Errorf("TopologyConnect failed: %w", err)
			}
		}
	}

	return nil
}
