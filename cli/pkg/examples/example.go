package examples

import (
	"embed"
	"errors"
	"os"
	"path"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

//go:embed all:embed
var examplesFS embed.FS

var ErrExampleNotFound = errors.New("example not found")

func List() ([]string, error) {
	entries, err := examplesFS.ReadDir("embed/dist")
	if err != nil {
		return nil, err
	}

	exampleList := make([]string, 0, len(entries))
	for _, entry := range entries {
		exampleList = append(exampleList, entry.Name())
	}

	return exampleList, nil
}

func Init(name string, projectDir string) error {
	examplePath := path.Join("embed", "dist", name)

	_, err := examplesFS.ReadDir(examplePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrExampleNotFound
		}
		return err
	}

	return fileutil.CopyEmbedDir(examplesFS, examplePath, projectDir)
}
