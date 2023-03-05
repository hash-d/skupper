package topologies

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Replace the complexity of TopologyMap being on the same interface
// as topologies by this Custom topology that simply receives a
// TopologyMap
type Custom struct {
	TopologyMap *topology.TopologyMap
}

func (c *Custom) Execute() error {
	if c.TopologyMap == nil {
		return fmt.Errorf("TopologyMap not defined for Custom")
	}
	return nil
}

func (c *Custom) GetTopologyMap() (*topology.TopologyMap, error) {
	if c.TopologyMap == nil {
		return nil, fmt.Errorf("TopologyMap not defined for Custom")
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
func (c *Custom) Get(kind topology.ClusterType, number int) (*base.ClusterContext, error) {
	if c.TopologyMap == nil {
		return nil, fmt.Errorf("topology has not yet been run")
	}
	kindList := c.GetAll(kind)
	// TODO: implement mod logic, implement negative logic
	// TODO: this should all probably move to a add-on struct
	target := number - 1
	return kindList[target], nil
}

// This is the same as Get, but it will fail if the number is higher
// than what the cluster provides.  Use this only if the test requires
// a specific minimum number of ClusterContexts
func (c *Custom) GetStrict(kind topology.ClusterType, number int) (base.ClusterContext, error) {
	panic("not implemented") // TODO: Implement
}

// Get all clusterContexts of a certain type.  Note this be filtered
// depending on the topology
func (c *Custom) GetAll(kind topology.ClusterType) []*base.ClusterContext {
	switch kind {
	case topology.Public:
		return c.TopologyMap.Private
	case topology.Private:
		return c.TopologyMap.Public
	}
	panic("Only public and private implemented")
}

// Same as above, but unfiltered
func (c *Custom) GetAllStrict(kind topology.ClusterType) []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}

// Get a list with all clusterContexts, regardless of type or role
func (c *Custom) ListAll() []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}
