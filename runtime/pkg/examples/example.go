package examples

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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

	// Copy example project to projectDir
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

func Unpack(name string) ([]fs.File, []string, error) {
	exampleDir := path.Join("embed", "dist", name)
	paths, err := doublestar.Glob(examplesFS, path.Join(exampleDir, "**"))
	if err != nil {
		return nil, nil, err
	}

	if len(paths) == 0 {
		return nil, nil, ErrExampleNotFound
	}

	filePaths := make([]string, len(paths))
	files := make([]fs.File, len(paths))
	for i, path := range paths {
		file, err := examplesFS.Open(path)
		if err != nil {
			return nil, nil, err
		}

		files[i] = file
		_, filePaths[i], _ = strings.Cut(paths[i], exampleDir)
	}

	return files, filePaths, nil
}
