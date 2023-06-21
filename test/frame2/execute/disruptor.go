package execute

import "github.com/skupperproject/skupper/test/utils/base"

// This interface should be used only on components that
// install skupper (such as SkupperInit).  It indicates to
// the upgrade disruptors that a step and/or namespace is
// a candidate for running skupper upgrade
type SkupperUpgradable interface {
	//
	SkupperUpgradable() *base.ClusterContext
}

type SkupperCliPathSetter interface {
	SetSkupperCliPath(path string, env []string)
}
