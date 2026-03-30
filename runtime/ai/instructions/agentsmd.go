package instructions

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"gopkg.in/yaml.v3"
)

// InitAgentsMD generates tool-agnostic AGENTS.md instruction files from Rill instruction files.
// The entry point is written to /AGENTS.md.
// All other instructions (including development.md) are written as skills to /.agents/skills/<name>/SKILL.md.
// MCP server config is written to /.mcp.json.
// If force is false, it skips files that already exist.
// If force is true, it overwrites any existing files.
func InitAgentsMD(ctx context.Context, repo drivers.RepoStore, force bool) error {
	// Load all instruction files with External=true for external editor use
	instructions, err := LoadAll(Options{External: true})
	if err != nil {
		return fmt.Errorf("failed to load instructions: %w", err)
	}

	// Convert and write each instruction file
	for path, inst := range instructions {
		outputPath, content := convertToAgentsMDFile(path, inst)

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

	// Write MCP server config
	err = writeMCPConfig(ctx, repo, force, "/.mcp.json", map[string]any{
		"type": "http",
		"url":  "http://localhost:9009/mcp",
	})
	if err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}

// convertToAgentsMDFile transforms a Rill instruction to AGENTS.md format.
// AGENTS.md becomes the main /AGENTS.md file.
// Other files become skills at /.agents/skills/<name>/SKILL.md.
func convertToAgentsMDFile(path string, inst *Instruction) (outputPath, content string) {
	// AGENTS.md becomes the main AGENTS.md file (no front matter needed)
	if path == "AGENTS.md" {
		return "/AGENTS.md", inst.Body
	}

	// Other files become skills
	name := fmt.Sprintf("rill-%s", strings.ReplaceAll(inst.Name, "_", "-"))
	outputPath = "/.agents/skills/" + name + "/SKILL.md"

	// Serialize front matter to YAML
	fmBytes, _ := yaml.Marshal(&skillFrontMatter{
		Name:        name,
		Description: inst.Description,
	})

	// Build final content with front matter
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(inst.Body)

	return outputPath, sb.String()
}
