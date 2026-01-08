package instructions

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed all:data
var dataFS embed.FS

// Instruction represents a parsed instruction file with front matter and body.
type Instruction struct {
	Name        string
	Description string
	Body        string
}

// Options configures how instruction files are loaded and rendered.
type Options struct {
	// External indicates whether the instructions are being loaded for external use (e.g., Claude Skills or Cursor rules) or internal use (e.g., Rill's own agents).
	External bool
}

// frontMatter represents the YAML front matter of an instruction file.
type frontMatter struct {
	Description string `yaml:"description"`
}

// Load loads a single instruction file by path (relative to the data directory).
// The path should include the file extension, e.g., "development.md" or "resources/model.md".
func Load(path string, opts Options) (*Instruction, error) {
	fullPath := filepath.Join("data", path)
	content, err := dataFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read instruction file %q: %w", path, err)
	}

	return parseInstruction(path, content, opts)
}

// LoadAll loads all instruction files from the data directory recursively.
// Returns a map of file paths (relative to data directory) to their parsed instructions.
func LoadAll(opts Options) (map[string]*Instruction, error) {
	instructions := make(map[string]*Instruction)

	err := fs.WalkDir(dataFS, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := dataFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read instruction file %q: %w", path, err)
		}

		instruction, err := parseInstruction(path, content, opts)
		if err != nil {
			return fmt.Errorf("failed to parse instruction file %q: %w", path, err)
		}

		// Store with path relative to data directory
		relPath := strings.TrimPrefix(path, "data/")
		instructions[relPath] = instruction

		return nil
	})
	if err != nil {
		return nil, err
	}

	return instructions, nil
}

// parseInstruction parses an instruction file's content, extracting front matter and applying templates to the body.
func parseInstruction(path string, content []byte, opts Options) (*Instruction, error) {
	// Extract name from path
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	// Parse front matter
	fm, body, err := parseFrontMatter(content)
	if err != nil {
		return nil, err
	}

	// Apply template to the body
	renderedBody, err := executeTemplate(body, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return &Instruction{
		Name:        name,
		Description: fm.Description,
		Body:        renderedBody,
	}, nil
}

// parseFrontMatter extracts YAML front matter from markdown content.
// Front matter is expected to be delimited by "---" at the start and end.
func parseFrontMatter(content []byte) (*frontMatter, string, error) {
	contentStr := strings.TrimSpace(string(content))

	// Check for front matter delimiter at the start
	if !strings.HasPrefix(contentStr, "---\n") && !strings.HasPrefix(contentStr, "---\r\n") {
		// No front matter, return empty front matter and full content as body
		return &frontMatter{}, contentStr, nil
	}

	// Find the closing delimiter
	// Skip the first "---\n" and find the next "---"
	rest := contentStr[4:] // Skip "---\n" or start of "---\r\n"
	if strings.HasPrefix(contentStr, "---\r\n") {
		rest = contentStr[5:]
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, "", fmt.Errorf("unclosed front matter: missing closing ---")
	}

	frontMatterContent := rest[:endIdx]
	body := rest[endIdx+4:] // Skip "\n---"

	// Handle optional newline after closing delimiter
	body = strings.TrimSpace(body)

	// Parse the front matter YAML
	var fm frontMatter
	if err := yaml.Unmarshal([]byte(frontMatterContent), &fm); err != nil {
		return nil, "", fmt.Errorf("failed to parse front matter YAML: %w", err)
	}

	return &fm, body, nil
}

// executeTemplate applies Go's template engine to the instruction body.
// Uses custom delimiters "{%" and "%}" to avoid conflicts with Go template syntax that may appear in example code within the instruction markdown.
func executeTemplate(body string, opts Options) (string, error) {
	tmpl, err := template.New("instruction").Delims("{%", "%}").Parse(body)
	if err != nil {
		return "", err
	}

	data := map[string]any{
		"external": opts.External,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
