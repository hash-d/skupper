package topologies

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2/topology"
)

// Replace the complexity of TopologyMap being on the same interface
// as topologies by this Custom topology that simply receives a
// TopologyMap
type Custom struct {
	TopologyMap *topology.TopologyMap

	contextHolder
}

func (c *Custom) Execute() error {
	if c.TopologyMap == nil {
		return fmt.Errorf("TopologyMap not defined for Custom")
	}
	c.contextHolder = contextHolder{TopologyMap: c.TopologyMap}
	return nil
}
