package execute

import (
	"regexp"
	"sort"
	"strconv"

	"github.com/skupperproject/skupper/test/utils/base"
)

// TODO these should all move to a skupper-specific thing

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

// An action that implements SkupperVersioner is expected
// to act differently for different Skupper versions.  This
// can be used, for example, when flags are added or removed
// to the cli.
//
// The action should have sub-actions that are version-specific.
// Whenever a new version changes something, a new sub action
// should be created, copying the code from the older one and
// adding the changes.  Avoid having a bunch of 'if version' on
// the code; just duplicate and change on the new, to keep it
// simple.
//
// The main action should then run the appropriate sub action
// depending on its assigned version.  If a flag changed default
// value from one version to the other, for example, it should
// make that change.
//
// SkupperVersioner and all other interfaces should be implemented
// on the main action only; let it configure the sub-actions on
// its Execute call and call them
type SkupperVersioner interface {
	SetSkupperVersion(version string)
	GetSkupperVersion() string
}

type SkupperVersionerDefault struct {
	SkupperVersion string
}

func (s *SkupperVersionerDefault) SetSkupperVersion(version string) {
	s.SkupperVersion = version
}

func (s SkupperVersionerDefault) GetSkupperVersion() string {
	return s.SkupperVersion
}

// Given a list of versions, WhichSkupperVersion will return the one that
// is more appropriate to be used, given its current SkupperVersion value.
//
// Namely:
//
//   - If its SkupperVersion is empty, always return empty, regardless of the
//     values of the candidates
//   - If its SkupperVersion is greater than all presented candidates, return
//     empty, indicating that the most recent version should be used
//   - If its SkupperVersion is lower than all presented candidates, return
//     the candidate with the lowest version
//   - If its SkupperVersion stands in between two versions, return the
//     lower version of the two
//
// The way SkupperVersioner is to be used, changes are always introduced
// on the sub action named after the version that introduces the change.  So,
// if something changed on 1.4, WhichSkupperVersion receives the candidates
// 1.2 and 1.4, and its current SkupperVersion is 1.3, it will return 1.2
// (that is, the version that 1.3 is compatible with, as it does not have the
// changes from 1.4).
func (s SkupperVersionerDefault) WhichSkupperVersion(candidates []string) string {
	// The action is configured to use the latest, so always return empty
	if s.SkupperVersion == "" || len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return VersionLessThan(candidates[i], candidates[j])
	})

	bestMatch := candidates[0]
	for _, item := range candidates {
		if item == s.SkupperVersion {
			return item
		}
		if !VersionLessThan(item, s.SkupperVersion) {
			return bestMatch
		} else {
			bestMatch = item
		}
	}

	return ""

}

// As the name implies, compares two versions in string form.  It expects the
// X.Y.Z-a-b-c format, where any items that are integers on both sides will
// be compared numerically
func VersionLessThan(version, than string) bool {

	// Compare X, Y, Z or something that's within the '-' part of the version
	compareItem := func(version, than []string, i int) (result bool, stop bool) {
		var iv, it int
		var sv, st string
		var errv, errt error
		var vPastEnd, tPastEnd bool

		// if one of the versions is shorter than i, we consider it
		// as the number zero
		if i < len(version) {
			sv = version[i]
			iv, errv = strconv.Atoi(version[i])
		} else {
			vPastEnd = true
		}
		if i < len(than) {
			st = than[i]
			it, errv = strconv.Atoi(than[i])
		} else {
			tPastEnd = true
		}

		// We reached i, and we have nothing else to compare; the versions
		// are the same
		if vPastEnd && tPastEnd {
			return false, true
		}
		if vPastEnd {
			if errt != nil {
				if it == 0 {
					// We can't be sure yet
					return false, false
				}
				// "than" more specified than "version", so we consider "version" to
				// be less than "than"
				return true, true
			}
		}
		if tPastEnd {
			if errv != nil {
				if iv == 0 {
					return false, false
				}
				// Similarly, "version" is more specified than "than"
				return false, true
			}
		}

		if errv == nil && errt == nil {
			// We have a direct int to int comparison
			stop = iv != it
			return iv < it, stop
		}

		// if we got here, at least one of the pieces is not an integer
		stop = sv != st
		return sv < st, stop

	}

	sv := regexp.MustCompile("[.-]").Split(version, -1)
	st := regexp.MustCompile("[.-]").Split(than, -1)

	for i := 0; 1 < 1000; i++ {
		result, stop := compareItem(sv, st, i)
		if stop {
			return result
		}
	}

	panic("This should not be reachable")
}
