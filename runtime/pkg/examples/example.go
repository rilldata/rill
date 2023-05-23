package examples

import (
	"embed"
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
)

//go:embed all:embed
var examplesFS embed.FS

var ErrExampleNotFound = errors.New("example not found")

type Example struct {
	Name        string
	Title       string
	Description string
}

func List() ([]Example, error) {
	entries, err := examplesFS.ReadDir("embed/dist")
	if err != nil {
		return nil, err
	}

	exampleList := make([]Example, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		rillYamlContents, err := examplesFS.ReadFile(filepath.Join("embed/dist", entry.Name(), "rill.yaml"))
		if err != nil {
			return nil, err
		}

		rillYaml, err := rillv1beta.ParseProjectConfig(rillYamlContents)
		if err != nil {
			return nil, err
		}

		exampleList = append(exampleList, Example{
			Name:        entry.Name(),
			Title:       rillYaml.Title,
			Description: rillYaml.Description,
		})
	}

	return exampleList, nil
}

func Unpack(name string) (fs.FS, error) {
	return fs.Sub(examplesFS, filepath.Join("embed", "dist", name))
}
