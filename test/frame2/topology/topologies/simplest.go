package topologies

import (
	"github.com/skupperproject/skupper/test/frame2/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

// It's the simplest Skupper topology you can get: prv1 connects
// to pub1.  That's it.
type Simplest struct {
	Name           string
	TestRunnerBase *base.ClusterTestRunnerBase

	ConsoleOnPublic  bool
	ConsoleOnPrivate bool

	// Add on
	*contextHolder

	Return *topology.TopologyMap
}

func (s *Simplest) Execute() error {

	pub1 := &topology.TopologyItem{
		Type:                topology.Public,
		EnableConsole:       s.ConsoleOnPublic,
		EnableFlowCollector: s.ConsoleOnPublic,
	}
	prv1 := &topology.TopologyItem{
		Type:                topology.Private,
		EnableConsole:       s.ConsoleOnPrivate,
		EnableFlowCollector: s.ConsoleOnPrivate,
		Connections: []*topology.TopologyItem{
			pub1,
		},
	}
	topoMap := []*topology.TopologyItem{
		pub1,
		prv1,
	}

	s.Return = &topology.TopologyMap{
		Name:           s.Name,
		TestRunnerBase: s.TestRunnerBase,
		Map:            topoMap,
	}

	s.contextHolder = &contextHolder{TopologyMap: s.Return}

	return nil
}
