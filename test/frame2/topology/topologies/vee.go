package topologies

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

// A topology in V shape; odd-numbered namespaces go on the
// left branch, even on the right branch.  Except for one
// additional namespace, that can be selected between pub
// and private, and connects both branches.
//
// The constant that pub1 and prv1 are the farthest apart
// possible on the topology is true within the 'left branch'
//
// Similarly, pub2 and prv2 will be the farthest apart on the
// right branch.
type V struct {
	Name           string
	TestRunnerBase *base.ClusterTestRunnerBase

	EmptyRight bool // If set, do not deploy Skupper or applications on the right branch
	VertixType topology.ClusterType

	//For the future
	// VertixConnectionClusterType // whether Vertix should connect to a pub or private cluster
	// Invert right // inverts the selection above for the right branch
	// NumPub, NumPrv. Allow segments

	Return *topology.TopologyMap
}

func (v *V) Execute() error {
	pub1 := &topology.TopologyItem{
		Type: topology.Public,
	}
	prv1 := &topology.TopologyItem{
		Type: topology.Private,
		Connections: []*topology.TopologyItem{
			pub1,
		},
	}
	pub2 := &topology.TopologyItem{
		Type: topology.Public,
	}
	prv2 := &topology.TopologyItem{
		Type: topology.Private,
		Connections: []*topology.TopologyItem{
			pub2,
		},
	}
	other := &topology.TopologyItem{
		Type: topology.Public,
		Connections: []*topology.TopologyItem{
			pub1,
			pub2,
		},
	}

	topoMap := []*topology.TopologyItem{
		pub1,
		prv1,
		pub2,
		prv2,
		other,
	}

	v.Return = &topology.TopologyMap{
		Name:           v.Name,
		TestRunnerBase: v.TestRunnerBase,
		Map:            topoMap,
	}

	return nil
}

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
func (v *V) Get(kind topology.ClusterType, number int) (*base.ClusterContext, error) {
	if v.Return == nil {
		return nil, fmt.Errorf("topology has not yet been run")
	}
	kindList := v.GetAll(kind)
	// TODO: implement mod logic, implement negative logic
	// TODO: this should all probably move to a add-on struct
	target := number - 1
	return kindList[target], nil
}

// This is the same as Get, but it will fail if the number is higher
// than what the cluster provides.  Use this only if the test requires
// a specific minimum number of ClusterContexts
func (v *V) GetStrict(kind topology.ClusterType, number int) (base.ClusterContext, error) {
	panic("not implemented") // TODO: Implement
}

// Get all clusterContexts of a certain type.  Note this be filtered
// depending on the topology
func (v *V) GetAll(kind topology.ClusterType) []*base.ClusterContext {
	switch kind {
	case topology.Public:
		return v.Return.Public
	case topology.Private:
		return v.Return.Private
	}
	panic("Only public and private implemented")
}

// Same as above, but unfiltered
func (v *V) GetAllStrict(kind topology.ClusterType) []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}

// Get a list with all clusterContexts, regardless of type or role
func (v *V) ListAll() []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}

func (v *V) GetTopologyMap() (*topology.TopologyMap, error) {
	if v.Return == nil {
		return nil, fmt.Errorf("topologyMap is nil; not yet run?")
	}
	return v.Return, nil
}
