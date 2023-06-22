package execute

import (
	"fmt"
	"log"
	"testing"

	"gotest.tools/assert"
)

func TestVersionLessOrEqualThan(t *testing.T) {

	table := []struct {
		version string
		than    string
		result  bool
	}{
		{
			version: "1.0.0",
			than:    "1.0.0",
			result:  false,
		}, {
			version: "1.0.0",
			than:    "1.1.0",
			result:  true,
		}, {
			version: "1.0.0",
			than:    "1.0.1",
			result:  true,
		}, {
			version: "1.0",
			than:    "1.0.1",
			result:  true,
		}, {
			version: "1.0",
			than:    "1.0.0",
			result:  false,
		}, {
			version: "1.0.0",
			than:    "1.0.0-something",
			result:  true,
		}, {
			version: "1.0.0-something",
			than:    "1.0.0",
			result:  false,
		}, {
			version: "1.0.0-a-1",
			than:    "1.0.0-z-1",
			result:  true,
		}, {
			version: "1.0.0-z-1",
			than:    "1.0.0-a-1",
			result:  false,
		}, {
			version: "1.0.0-something-1",
			than:    "1.0.0-something-1",
			result:  false,
		}, {
			version: "1.0.0-something-1",
			than:    "1.0.0-something-2",
			result:  true,
		}, {
			version: "1.0.0-something-2",
			than:    "1.0.0-something-1",
			result:  false,
		}, {
			version: "2",
			than:    "10",
			result:  true,
		}, {
			version: "10",
			than:    "2",
			result:  false,
		},
	}

	for i, item := range table {
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Logf("testing %q < %q: %t", item.version, item.than, item.result)
			assert.Assert(t, item.result == VersionLessThan(item.version, item.than))
		})
	}

}

func TestWhichSkupperVersion(t *testing.T) {
	table := []struct {
		candidates []string
		version    string
		result     string
	}{
		{
			version: "",
			result:  "",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "2",
			result:     "2",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "2.1",
			result:     "2",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "3",
			result:     "2",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "4",
			result:     "4",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "5",
			result:     "4",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "6",
			result:     "6",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "7",
			result:     "",
		}, {
			candidates: []string{"2", "4", "6"},
			version:    "1",
			result:     "2",
		},
	}

	for i, item := range table {
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			log.Printf(
				"Testing candidates %+v with version %q expecting result  %q",
				item.candidates,
				item.version,
				item.result,
			)
			s := SkupperVersionerDefault{
				SkupperVersion: item.version,
			}
			actual := s.WhichSkupperVersion(item.candidates)
			log.Printf("Got %q", actual)

			assert.Assert(t, actual == item.result)
		})
	}
}
