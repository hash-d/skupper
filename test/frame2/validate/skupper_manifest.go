package validate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/skupperproject/skupper/test/frame2"
)

type SkupperManifestContentImage struct {
	Name       string
	Repository string
}

type SkupperManifestContent struct {
	Images []SkupperManifestContentImage
}

// SkupperManifest returns the content of the requested manifest.json
// file as a SkupperManifestContent.
//
// If data is provided in Expected, it also checks that all of its items
// match the actual file's contents.
//
// The check is only on the actual images, as strings, including the tags.
// It has no intelligence to add or remove :latest from a tag, for example.
//
// The Repository field is not used in this verification.
type SkupperManifest struct {

	// Path to the manifest.json file; if not provided, it will be
	// searched first on the test root, then on the source root.
	Path string

	SkipComparison bool

	Expected SkupperManifestContent
	Result   *SkupperManifestContent

	frame2.DefaultRunDealer
	frame2.Log
}

func (m SkupperManifest) Validate() error {
	manifestPath := m.Path
	alternates := []string{
		"./manifest.json",
		path.Join(frame2.SourceRoot(), "manifest.json"),
	}
	if manifestPath == "" {
		for _, alternate := range alternates {
			if _, err := os.Stat(alternate); err == nil {
				manifestPath = alternate
				break
			}
		}
	}
	if manifestPath == "" {
		return fmt.Errorf("SkupperManifest: no path to manifest.json found, and none found on default locations")
	}

	var manifestBytes []byte
	manifestBytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("SkupperManifest: could not read file %q: %w", manifestPath, err)
	}

	m.Result = &SkupperManifestContent{}
	err = json.Unmarshal(manifestBytes, m.Result)
	if err != nil {
		return fmt.Errorf("SkupperManifest: could not unmarshal %q: %w", manifestPath, err)
	}

	if m.SkipComparison {
		m.Log.Printf("SkupperManifest>: Skipping comparison per configuration")
		return nil
	}

	for _, expected := range m.Expected.Images {
		var found bool
		for _, actual := range m.Result.Images {
			if expected.Name == actual.Name {
				m.Log.Printf("Image %q matched on %q (repo %q)", actual.Name, manifestPath, actual.Repository)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Image %q did not match any items from %q", expected.Name, manifestPath)
		}

	}

	return nil
}
