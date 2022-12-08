package environment

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/frame2/deploy"
	"github.com/skupperproject/skupper/test/frame2/topology"
)

// A Hello World deployment on pub1 (frontend) and prv1 (backend),
// on an N topology.
//
// Useful for the simplest multiple link testing.
//
// See topology.N for details on the topology.
type HelloWorldN struct {
}

// A Hello World deployment, with configurations.  For simpler
// alternatives, see:
//
// - environment.HelloWorldSimple
// - environment.HelloWorldN
// - ...
// - environment.HelloWorldPlatform is special. It will use
//   whatever topology the current test is asking for, if
//   possible
//
type HelloWorld struct {
	TopologyMap topology.TopologyMap
}

func (hw HelloWorld) Execute() error {
	log.Printf("environment.HelloWorld")
	log.Printf("create topology")
	topo := topology.Topology{
		TopologyMap: hw.TopologyMap,
	}
	err := topo.Execute()
	if err != nil {
		return fmt.Errorf("HelloWorld failed to create topology: %w", err)
	}

	deployHW := deploy.HelloWorld{
		Topology: topo,
	}
	err = deployHW.Execute()
	if err != nil {
		return fmt.Errorf("HelloWorld failed deployment: %w", err)
	}
	log.Printf("TODO deploy app")
	return nil
}
