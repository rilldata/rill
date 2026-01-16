package instructions

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"gopkg.in/yaml.v3"
)

// InitCursorRules generates Cursor rules files from Rill instruction files.
// The rules are written to .cursor/rules/ in the repository.
// If force is false, it skips generation if .cursor/rules/ already exists.
// If force is true, it overwrites any existing files.
func InitCursorRules(ctx context.Context, repo drivers.RepoStore, force bool) error {
	// Load all instruction files with External=true for external editor use
	instructions, err := LoadAll(Options{External: true})
	if err != nil {
		return fmt.Errorf("failed to load instructions: %w", err)
	}

	// Convert and write each instruction file
	for path, inst := range instructions {
		outputPath, content := convertToCursorRule(path, inst)

		if !force {
			_, err := repo.Stat(ctx, outputPath)
			if err == nil {
				// File exists, skip
				continue
			}
		}

		err = repo.Put(ctx, outputPath, strings.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to write %q: %w", outputPath, err)
		}
	}

	return nil
}

// cursorFrontMatter represents the YAML front matter for Cursor rules files.
type cursorFrontMatter struct {
	Description string `yaml:"description"`
	AlwaysApply bool   `yaml:"alwaysApply"`
}

// convertToCursorRule transforms a Rill instruction to Cursor rule format.
func convertToCursorRule(path string, inst *Instruction) (outputPath, content string) {
	// Determine output path: .md -> .mdc, under .cursor/rules/
	outputPath = strings.TrimSuffix(path, ".md") + ".mdc"
	outputPath = "/.cursor/rules/" + outputPath

	// Serialize front matter to YAML
	fmBytes, _ := yaml.Marshal(&cursorFrontMatter{
		Description: inst.Description,
		AlwaysApply: path == "development.md",
	})

	// Build final content
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(inst.Body)

	return outputPath, sb.String()
}
