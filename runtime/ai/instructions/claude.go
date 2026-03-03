package instructions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"gopkg.in/yaml.v3"
)

// InitClaudeCode generates Claude Code instruction files from Rill instruction files.
// The main instructions are written to .claude/CLAUDE.md.
// Resource-specific instructions are written as skills to .claude/skills/<name>/SKILL.md.
// Skills are loaded on-demand when invoked, keeping the context lean.
// If force is false, it skips files that already exist.
// If force is true, it overwrites any existing files.
func InitClaudeCode(ctx context.Context, repo drivers.RepoStore, force bool) error {
	// Load all instruction files with External=true for external editor use
	instructions, err := LoadAll(Options{External: true})
	if err != nil {
		return fmt.Errorf("failed to load instructions: %w", err)
	}

	// Convert and write each instruction file
	for path, inst := range instructions {
		outputPath, content := convertToClaudeFile(path, inst)

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

	// Write MCP server config (Claude uses .mcp.json file at root, not under .claude)
	err = writeMCPConfig(ctx, repo, force, "/.mcp.json", map[string]any{
		"type": "http",
		"url":  "http://localhost:9009/mcp",
	})
	if err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}

// skillFrontMatter represents the YAML front matter for Claude Code SKILL.md files.
type skillFrontMatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// convertToClaudeFile transforms a Rill instruction to Claude Code format.
// development.md becomes the main .claude/CLAUDE.md file.
// Resource files become skills at .claude/skills/<name>/SKILL.md.
func convertToClaudeFile(path string, inst *Instruction) (outputPath, content string) {
	// development.md becomes the main CLAUDE.md file (no front matter needed)
	if path == "development.md" {
		return "/.claude/CLAUDE.md", inst.Body
	}

	// Other files become skills
	name := fmt.Sprintf("rill-%s", strings.ReplaceAll(inst.Name, "_", "-"))
	outputPath = "/.claude/skills/" + name + "/SKILL.md"

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

// mcpServerName is the name used for the Rill MCP server in editor configs.
const mcpServerName = "rill-developer"

// writeMCPConfig reads an existing MCP config file (if any), adds or updates
// the "rill" server entry, and writes the result back. If force is false and
// the "rill" entry already exists, it is left unchanged.
func writeMCPConfig(ctx context.Context, repo drivers.RepoStore, force bool, path string, serverConfig map[string]any) error {
	// Try to read and parse the existing config
	var cfg struct {
		MCPServers map[string]any `json:"mcpServers"`
	}
	existing, err := repo.Get(ctx, path)
	if err == nil && existing != "" {
		_ = json.Unmarshal([]byte(existing), &cfg)
	}

	// If not forcing and the entry already exists, skip
	if !force {
		if _, ok := cfg.MCPServers[mcpServerName]; ok {
			return nil
		}
	}

	// Update the config with the new server entry
	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]any)
	}
	cfg.MCPServers[mcpServerName] = serverConfig

	// Marshal and write back the updated config
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}
	return repo.Put(ctx, path, strings.NewReader(string(data)+"\n"))
}
