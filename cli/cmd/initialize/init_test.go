package initialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddCursorRules_Table(t *testing.T) {
	tcs := []struct {
		name          string
		force         bool
		setupFiles    map[string]string
		expectEqual   map[string]string
		expectContain map[string]string
	}{
		{
			name:  "skips existing when not forced",
			force: false,
			setupFiles: map[string]string{
				"code-style.md": "OLD",
			},
			expectEqual: map[string]string{
				"code-style.md": "OLD",
			},
			expectContain: map[string]string{
				"project-structure.md": "# Project Structure",
				"best-practices.md":    "# Best Practices",
			},
		},
		{
			name:  "force overwrites all",
			force: true,
			setupFiles: map[string]string{
				"code-style.md":        "OLD",
				"project-structure.md": "OLD",
				"best-practices.md":    "OLD",
			},
			expectContain: map[string]string{
				"code-style.md":        "# Code Style",
				"project-structure.md": "# Project Structure",
				"best-practices.md":    "# Best Practices",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			cwd, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(cwd) }()

			require.NoError(t, os.Chdir(tmp))

			// create .cursor/rules and pre-populate any requested files
			require.NoError(t, os.MkdirAll(filepath.Join(".cursor", "rules"), 0o755))
			for rel, content := range tc.setupFiles {
				require.NoError(t, os.WriteFile(filepath.Join(".cursor", "rules", rel), []byte(content), 0o644))
			}

			err = addCursorRules(tc.force, tmp)
			require.NoError(t, err)

			for f, want := range tc.expectEqual {
				p := filepath.Join(".cursor", "rules", f)
				b, err := os.ReadFile(p)
				require.NoError(t, err)
				require.Equal(t, want, string(b))
			}

			for f, want := range tc.expectContain {
				p := filepath.Join(".cursor", "rules", f)
				b, err := os.ReadFile(p)
				require.NoError(t, err)
				require.Contains(t, string(b), want)
			}
		})
	}
}
