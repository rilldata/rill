package examples

import (
	"embed"
	"errors"
	"os"
	"path"
	"path/filepath"

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

func Init(name, projectDir string) error {
	examplePath := path.Join("embed", "dist", name)

	_, err := examplesFS.ReadDir(examplePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrExampleNotFound
		}
		return err
	}

	// We want to append to .gitignore, not override it.
	// Cache it here.
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	prevGitignore, _ := os.ReadFile(gitignorePath)

	err = fileutil.CopyEmbedDir(examplesFS, examplePath, projectDir)
	if err != nil {
		return err
	}

	// Fix up gitignore
	if len(prevGitignore) != 0 {
		newGitignore, _ := os.ReadFile(gitignorePath)
		gitignore := string(prevGitignore) + "\n" + string(newGitignore)
		err := os.WriteFile(gitignorePath, []byte(gitignore), os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
