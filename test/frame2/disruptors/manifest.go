package disruptors

import (
	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/frame2/validate"
)

type SkipManifestCheck struct {
}

func (s SkipManifestCheck) DisruptorEnvValue() string {
	return "SKIP_MANIFEST_CHECK"
}

func (s *SkipManifestCheck) Inspect(step *frame2.Step, phase *frame2.Phase) {
	for _, v := range step.GetValidators() {
		if v, ok := v.(*validate.SkupperManifest); ok {
			v.SkipComparison = true
		}
	}
}
