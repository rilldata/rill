package instructions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Test loading a single instruction file
	inst, err := Load("development.md", Options{External: false})
	require.NoError(t, err)
	require.NotNil(t, inst)

	require.Equal(t, "Developing a Rill Project", inst.Name)
	require.Equal(t, "A general introduction to Rill's concepts, resource types, and project development process", inst.Description)
	require.NotEmpty(t, inst.Body)
	require.Contains(t, inst.Body, "# Instructions for developing a Rill project")
}

func TestLoadNested(t *testing.T) {
	// Test loading a nested instruction file
	inst, err := Load("resources/model.md", Options{External: false})
	require.NoError(t, err)
	require.NotNil(t, inst)
	require.NotEmpty(t, inst.Body)
}

func TestLoadNotFound(t *testing.T) {
	// Test loading a non-existent file
	_, err := Load("nonexistent.md", Options{External: false})
	require.Error(t, err)
}

func TestLoadAll(t *testing.T) {
	// Test loading all instruction files
	instructions, err := LoadAll(Options{External: false})
	require.NoError(t, err)
	require.NotEmpty(t, instructions)

	// Check that development.md is included
	dev, ok := instructions["development.md"]
	require.True(t, ok, "development.md should be loaded")
	require.Equal(t, "Developing a Rill Project", dev.Name)

	// Check that nested files are included
	_, ok = instructions["resources/model.md"]
	require.True(t, ok, "resources/model.md should be loaded")
}

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantName string
		wantDesc string
		wantBody string
		wantErr  bool
	}{
		{
			name: "valid front matter",
			content: `---
name: Test Name
description: Test Description
---

Body content here.`,
			wantName: "Test Name",
			wantDesc: "Test Description",
			wantBody: "\nBody content here.",
		},
		{
			name:     "no front matter",
			content:  "Just body content",
			wantName: "",
			wantDesc: "",
			wantBody: "Just body content",
		},
		{
			name: "unclosed front matter",
			content: `---
name: Test
description: Test`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := parseFrontMatter([]byte(tt.content))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantName, fm.Name)
			require.Equal(t, tt.wantDesc, fm.Description)
			require.Equal(t, tt.wantBody, body)
		})
	}
}

func TestTemplateExecution(t *testing.T) {
	// Uses custom delimiters {% %} to avoid conflicts with Go template
	// syntax in example code
	content := `---
name: Template Test
description: Testing templates
---

{%if .external %}External mode{% else %}Internal mode{% end %}`

	// Test with External = false
	inst, err := parseInstruction([]byte(content), Options{External: false})
	require.NoError(t, err)
	require.Contains(t, inst.Body, "Internal mode")
	require.NotContains(t, inst.Body, "External mode")

	// Test with External = true
	inst, err = parseInstruction([]byte(content), Options{External: true})
	require.NoError(t, err)
	require.Contains(t, inst.Body, "External mode")
	require.NotContains(t, inst.Body, "Internal mode")
}

func TestGoTemplateInExamplesPreserved(t *testing.T) {
	// Standard Go template syntax in example code should be preserved
	content := `---
name: Example Test
description: Testing that examples are preserved
---

Here is an example:
` + "```yaml" + `
sql: SELECT * FROM {{ ref "events" }}
` + "```"

	inst, err := parseInstruction([]byte(content), Options{External: false})
	require.NoError(t, err)
	require.Contains(t, inst.Body, `{{ ref "events" }}`)
}
