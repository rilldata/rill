package parser

import (
	"context"
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

//go:embed data/cursor/*.mdc
var cursorTemplatesFS embed.FS

// InitCursorRules adds Cursor AI rules to a Rill project
func InitCursorRules(ctx context.Context, repo drivers.RepoStore, force bool) error {
	// Walk the embedded data/cursor directory and copy files into .cursor/rules
	return fs.WalkDir(cursorTemplatesFS, "data/cursor", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Read the template content
		contentB, err := cursorTemplatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Determine the target path in .cursor/rules
		name := filepath.Base(path)
		targetPath := filepath.Join(".cursor", "rules", name)

		// Check if file already exists
		if !force {
			existing, _ := repo.Get(ctx, targetPath)
			if existing != "" {
				// File exists and force is not set, skip
				return nil
			}
		}

		// Write the file
		return repo.Put(ctx, targetPath, strings.NewReader(string(contentB)))
	})
}
