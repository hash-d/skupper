package topology

import "github.com/skupperproject/skupper/test/utils/base"

// Two pub, two private.  Connections always from prv to pub
//
// prv1 has two outgoing links; pub2 has two incoming links
//
//    pub1 pub2
//     |  / |     ^
//     | /  |     |   Connection direction
//    prv1 prv2
//
// Good for minimal multiple link testing
//
// TODO: change this topology, so that:
//
// - pub1 and prv1 have one link,
// - pub2 and prv2 have one link
//
// It will be easier to think about then this way.  The
// topology would then be:
//
//    pub2 pub1
//     | \  |     ^
//     |  \ |     |   Connection direction
//    prv1 prv2
//
// Also, pub1 and prv1 will be on the ends of the topology, which
// is the normal thing.
//
// TODO above
//
type N struct {
	Name           string
	TestRunnerBase *base.ClusterTestRunnerBase

	Return *TopologyMap
}

func (n *N) Execute() error {

	pub1 := &TopologyItem{
		Type: Public,
	}
	pub2 := &TopologyItem{
		Type: Public,
	}

	prv1 := &TopologyItem{
		Type: Private,
		Connections: []*TopologyItem{
			pub1,
			pub2,
		},
	}
	prv2 := &TopologyItem{
		Type: Private,
		Connections: []*TopologyItem{
			pub2,
		},
	}

	topoMap := []*TopologyItem{
		pub1,
		pub2,
		prv1,
		prv2,
	}

	n.Return = &TopologyMap{
		Name:           n.Name,
		TestRunnerBase: n.TestRunnerBase,
		Map:            topoMap,
	}

	return nil
}
