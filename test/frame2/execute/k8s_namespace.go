package execute

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type TestRunnerCreateNamespace struct {
	Namespace    base.ClusterContextPromise
	AutoTearDown bool
}

func (trcn TestRunnerCreateNamespace) Execute() error {
	log.Printf("TestRunnerCreateNamespace")
	cluster, err := trcn.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("TestRunnerCreateNamespace failed to create namespace from promise: %w", err)
	}

	log.Printf("Creating namespace %v", cluster.Namespace)

	err = cluster.CreateNamespace()
	if err != nil {
		return fmt.Errorf(
			"TestRunnerCreateNamespace failed to create namespace %q: %w",
			cluster.Namespace, err,
		)
	}

	return nil
}

func (trcn TestRunnerCreateNamespace) Teardown() frame2.Executor {
	if trcn.AutoTearDown {
		return TestRunnerDeleteNamespace{
			Namespace: trcn.Namespace,
		}
	}
	return nil
}

type TestRunnerDeleteNamespace struct {
	Namespace base.ClusterContextPromise
}

func (trdn TestRunnerDeleteNamespace) Execute() error {
	cluster, err := trdn.Namespace.Satisfy()
	if err != nil {
		return fmt.Errorf("TestRunnerCreateNamespace failed to delete namespace from promise: %w", err)
	}

	err = cluster.DeleteNamespace()
	if err != nil {
		return fmt.Errorf(
			"TestRunnerCreateNamespace failed to delete namespace %q: %w",
			cluster.Namespace, err,
		)
	}
	return nil
}
