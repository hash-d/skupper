package validate

import (
	"fmt"
	"testing"

	"github.com/skupperproject/skupper/test/frame2"
)

func TestSkupperManifest(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}

	expected := []SkupperManifestContentImage{
		{
			Name:       "quay.io/skupper/skupper-router:main",
			Repository: "https://github.com/skupperproject/skupper-router",
		},
		{
			Name:       "quay.io/skupper/service-controller:master",
			Repository: "https://github.com/skupperproject/skupper",
		},
		{
			Name:       "quay.io/skupper/config-sync:master",
			Repository: "https://github.com/skupperproject/skupper",
		},
		{
			Name:       "quay.io/skupper/flow-collector:master",
			Repository: "https://github.com/skupperproject/skupper",
		},
		{
			Name:       "quay.io/prometheus/prometheus:v2.42.0",
			Repository: "",
		},
	}

	for _, e := range expected {
		individualPhase := frame2.Phase{
			Runner: r,
			Doc:    fmt.Sprintf("Checks that %q is being checked individually, and also for error", e.Repository),
			MainSteps: []frame2.Step{
				{
					Doc: "Positive check",
					Validator: &SkupperManifest{
						Path: "testdata/manifest.json",
						Expected: SkupperManifestContent{
							Images: []SkupperManifestContentImage{
								{
									Name:       e.Name,
									Repository: e.Repository,
								},
							},
						},
					},
				}, {
					// Today, this is overkill, as we do not check Repository.  In practice, it checks that
					// :noexpected many times, with no additional checks
					Doc: "Negative check",
					Validator: &SkupperManifest{
						Path: "testdata/manifest.json",
						Expected: SkupperManifestContent{
							Images: []SkupperManifestContentImage{
								{
									Name:       "quay.io/skupper/skupper-router:notexpected",
									Repository: e.Repository,
								},
							},
						},
					},
					ExpectError: true,
				},
			},
		}

		individualPhase.Run()
	}
}
