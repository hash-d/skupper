package topology

import (
	"errors"
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Any topology needs to implement this interface.
//
// Its methods allow for components from frame2.deploy or tests to inquire
// about the topology's members, and deal with them without direct access to
// the TopologyMap.
type Basic interface {
	frame2.Executor

	GetTopologyMap() (*TopologyMap, error)

	// Return a ClusterContext of the given type and number.
	//
	// Negative numbers count from the end.  So, Get for -1 will return
	// the clusterContext with the greatest number of that type.
	//
	// Attention that for some types of topologies (suc as TwoBranched)
	// only part of the clustercontexts may be considered (for example,
	// only the left branch)
	//
	// The number divided against number of contexts of that type on
	// the topology, and the remainder will be used.  That allows for
	// tests that usually run with several namespace to run also with
	// a smaller number.  For example, on a cluster with 4 private
	// cluster, a request for number 6 will actually return number 2
	Get(kind ClusterType, number int) (*base.ClusterContext, error)

	// This is the same as Get, but it will fail if the number is higher
	// than what the cluster provides.  Use this only if the test requires
	// a specific minimum number of ClusterContexts
	GetStrict(kind ClusterType, number int) (base.ClusterContext, error)

	// Get all clusterContexts of a certain type.  Note this be filtered
	// depending on the topology
	GetAll(kind ClusterType) []*base.ClusterContext

	// Same as above, but unfiltered
	GetAllStrict(kind ClusterType) []base.ClusterContext

	// Get a list with all clusterContexts, regardless of type or role
	ListAll() []base.ClusterContext
}

// This represents a Topology which starts with a main branch and a secondary
// branch, connected by a Vertex.
//
// It can be used, for example, to test migrations.   The application and
// Skupper are initially installed on the Left branch, and the test moves it
// to the Right branch.
type TwoBranched interface {
	Basic

	// Same as Basic.Get(), but specifically on the left branch
	GetLeft(kind ClusterType, number int) (*base.ClusterContext, error)

	// Same as Basic.Get(), but specifically on the right branch
	GetRight(kind ClusterType, number int) (*base.ClusterContext, error)

	// Get the ClusterContext that connects the two branches
	GetVertex() (*base.ClusterContext, error)
}

// The type of cluster:
//
// - Public
// - Private
// - DMZ
//
// Currently, only the first two are implemented
type ClusterType int

const (
	Public ClusterType = iota
	Private
	DMZ
)

// TopoMap: receives

// A TopologyItem represents a skupper instalation on a namespace.
// The connections are outgoing links to other TopologyItems (or:
// to other Skupper installations)
//
// Once a TopologyItem has been realized by running its TopologyMap,
// the respective ClusterContext will be assigned
type TopologyItem struct {
	Type                  ClusterType
	Connections           []*TopologyItem
	SkipNamespaceCreation bool
	SkipSkupperDeploy     bool
	//SkipApplicationDeploy bool TODO

	ClusterContext *base.ClusterContext

	// TODO: need to add SkupperInstall configuration for the
	//       topology items, so site-specific configurations
	//       can be done (such as activating the console)
	EnableConsole bool
	// TODO
	EnableFlowCollector bool
}

// TopologyMap receives a list of TopologyItem that describe the topology.
//
// When executed, it creates the required ClusterContexts and returns three items:
//
// - A list of private clusterContexts
// - A list of public  clusterContexts
// - A go map from TopologyItem to ClusterContext
//
// These ClusterContexts do not yet refer to existing namespaces: that, along
// with Skupper installation and creation of the links is done by Topology and
// TopologyConnect.
//
// In general, tests should not use a TopologyMap as an executor.  Instead,
// just define it on a Topology, which will execute it.
//
// clients should keep a reference to a TopologyMap to
// get their output
type TopologyMap struct {
	// This will become the prefix on the name for the namespaces created
	Name string

	// All namespaces are created by a base.ClusterTestRunnerBase.
	TestRunnerBase *base.ClusterTestRunnerBase

	// Input
	Map []*TopologyItem

	// Output
	Private []*base.ClusterContext
	Public  []*base.ClusterContext

	GeneratedMap map[*TopologyItem]*base.ClusterContext
}

// Creates the ClusterContext items based on the provided map
//
// The actual namespaces are not yet created on this step.  Give the TopologyMap to a
// TopologyBuild to create them (and everything else)
//
// TODO: Validate: check for duplicates, disconnected items, etc (but allow to skip validation)
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
			item.ClusterContext = newContext
		case Private:
			countPrivate++
			newContext, err := tm.TestRunnerBase.GetPrivateContext(countPrivate)
			if err != nil {
				return fmt.Errorf("TopologyMap failed to get public context %d: %w", countPrivate, err)
			}
			tm.Private = append(tm.Private, newContext)
			tm.GeneratedMap[item] = newContext
			item.ClusterContext = newContext
		default:
			return errors.New("TopologyMap: Only Public and Private implemented")
		}
	}

	return nil
}

// TODO: Not yet implemented
type TopologyValidator struct {
	TopologyMap
}

func (tv TopologyValidator) Execute() error {
	log.Printf("TopologyValidator not yet implemented")
	return nil
}

// Based on a Topology, create the VAN:
//
// - Create the namespaces/ClusterContexts
// - Install Skupper
// - Create the links between the nodes.
//
// This ties together Topology, TopologyConnect
// and other items
type TopologyBuild struct {
	Topology     *Basic
	AutoTearDown bool

	// TODO Remove this?
	teardowns []frame2.Executor

	frame2.DefaultRunDealer
}

func (t *TopologyBuild) Execute() error {

	steps := frame2.Phase{
		Runner: t.GetRunner(),
		Doc:    "Create the topology",
		MainSteps: []frame2.Step{
			{
				Modify: *t.Topology,
			},
		},
	}
	steps.Execute()

	tm, err := (*t.Topology).GetTopologyMap()
	if err != nil {
		return fmt.Errorf("failed to get topologyMap: %w", err)
	}

	// Execute the TopologyMap; create the ClusterContext items
	buildTopologyMap := frame2.Phase{
		Runner: t.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: tm,
			},
		},
	}
	buildTopologyMap.Run()
	log.Printf("Generated TopologyMap: %+v", tm)

	log.Printf("Creating namespaces and installing Skupper")
	for topoItem, context := range tm.GeneratedMap {

		createAndInstall := frame2.Phase{
			Runner: t.Runner,
			Doc:    "Create namespaces and install Skupper",
			Setup: []frame2.Step{
				{
					Modify: execute.TestRunnerCreateNamespace{
						Namespace:    context,
						AutoTearDown: t.AutoTearDown,
					},
					SkipWhen: topoItem.SkipNamespaceCreation,
				}, {
					Modify: &execute.SkupperInstallSimple{
						Namespace:     context,
						EnableConsole: topoItem.EnableConsole,
					},
					SkipWhen: topoItem.SkipNamespaceCreation || topoItem.SkipSkupperDeploy,
				},
			},
		}
		createAndInstall.Execute()
	}

	connectSteps := frame2.Phase{
		Runner: t.Runner,
		Setup: []frame2.Step{
			{
				Modify: TopologyConnect{
					TopologyMap: *tm,
				},
			},
		},
	}
	connectSteps.Execute()

	return nil
}

// TODO Perhaps change the frame2.TearDowner interface to return a []frame2.Executor, instead, so a single
// call may return several, and have them run by the Runner?
func (t TopologyBuild) TearDown() frame2.Executor {
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
	// TODO: add some filters and run only one part of the topology
	// 	 (allow for late runs)
}

// Assumes that the namespaces are already created, and Skupper installed on all
// namespaces that will create or receive links
func (tc TopologyConnect) Execute() error {

	for from, ctx := range tc.TopologyMap.GeneratedMap {
		if from.SkipNamespaceCreation || from.SkipSkupperDeploy {
			continue
		}
		for _, to := range from.Connections {
			pivot := tc.TopologyMap.GeneratedMap[to]
			if to.SkipNamespaceCreation || to.SkipSkupperDeploy {
				continue
			}
			connName := fmt.Sprintf("%v-to-%v", ctx.Namespace, pivot.Namespace)
			log.Printf("TopologyConnect creating connection %v", connName)
			err := execute.SkupperConnect{
				Name: connName,
				From: ctx,
				To:   pivot,
			}.Execute()
			if err != nil {
				return fmt.Errorf("TopologyConnect failed: %w", err)
			}
		}
	}

	return nil
}
