package topologies

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

// This is an add-on for the topologies in this package.  When embedded into
// a topology struct, it will provide the functions that implement
// topology.Basic.
//
// Take note, however, to have its TopologyMap match the topology's (when
// setting it up, and in case that value is changed on the topology for
// any reason)
type contextHolder struct {
	TopologyMap *topology.TopologyMap
}

//func (c *ContextHolder) Execute() error {
//	panic("not implemented") // TODO: Implement
//}

func (c *contextHolder) GetTopologyMap() (*topology.TopologyMap, error) {
	if c.TopologyMap == nil {
		return nil, fmt.Errorf("ContextHolder: no TopologyMap defined")
	}
	return c.TopologyMap, nil
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
func (c *contextHolder) Get(kind topology.ClusterType, number int) (*base.ClusterContext, error) {
	if c.TopologyMap == nil {
		return nil, fmt.Errorf("topology has not yet been run")
	}
	kindList := c.GetAll(kind)
	// TODO: implement mod logic, implement negative logic
	// TODO: this should all probably move to a add-on struct
	if len(kindList) == 0 {
		return nil, fmt.Errorf("no clusterContext of type %v on the topology", kind)
	}
	var target int
	if number < 0 {
		target = len(kindList) + (number-1)%len(kindList) - 1
	} else {
		target = (number - 1) % len(kindList)
	}
	return kindList[target], nil
}

// This is the same as Get, but it will fail if the number is higher
// than what the cluster provides.  Use this only if the test requires
// a specific minimum number of ClusterContexts
func (c *contextHolder) GetStrict(kind topology.ClusterType, number int) (base.ClusterContext, error) {
	panic("not implemented") // TODO: Implement
}

// Get all clusterContexts of a certain type.  Note this be filtered
// depending on the topology
func (c *contextHolder) GetAll(kind topology.ClusterType) []*base.ClusterContext {
	switch kind {
	case topology.Public:
		return c.TopologyMap.Public
	case topology.Private:
		return c.TopologyMap.Private
	}
	panic("Only public and private implemented")

}

// Same as above, but unfiltered
func (c *contextHolder) GetAllStrict(kind topology.ClusterType) []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}

// Get a list with all clusterContexts, regardless of type or role
func (c *contextHolder) ListAll() []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}
