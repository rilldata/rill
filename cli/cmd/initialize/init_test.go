package initialize

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/stretchr/testify/require"
)

func TestInitCursorRulesCreatesAllFilesWhenNoneExist(t *testing.T) {
	tmp := t.TempDir()
	repo, instanceID, err := cmdutil.RepoForProjectPath(tmp)
	require.NoError(t, err)

	ctx := context.Background()
	err = parser.InitEmpty(ctx, repo, instanceID, "Test Project", "duckdb")
	require.NoError(t, err)

	// Ensure .cursor doesn't exist yet
	cursorFiles, _ := repo.Get(ctx, ".cursor/rules/code-style.mdc")
	require.Empty(t, cursorFiles)

	err = parser.InitCursorRules(ctx, repo, false)
	require.NoError(t, err)

	// Both template files should be created and contain expected headers
	files := map[string]string{
		".cursor/rules/code-style.mdc":        "# Rill Code Style",
		".cursor/rules/project-structure.mdc": "# Rill Project Structure",
	}

	for filePath, expectedContent := range files {
		contents, err := repo.Get(ctx, filePath)
		require.NoError(t, err)
		require.Contains(t, contents, expectedContent)
	}
}

func TestInitCursorRulesDoesNotOverwriteWithoutForce(t *testing.T) {
	tmp := t.TempDir()

	repo, instanceID, err := cmdutil.RepoForProjectPath(tmp)
	require.NoError(t, err)

	ctx := context.Background()
	err = parser.InitEmpty(ctx, repo, instanceID, "Test Project", "duckdb")
	require.NoError(t, err)

	// Create a cursor rule file with custom content using filesystem
	customContent := "# My Custom Rules\n"
	cursorDir := filepath.Join(tmp, ".cursor", "rules")
	err = os.MkdirAll(cursorDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(cursorDir, "code-style.mdc"), []byte(customContent), 0o644)
	require.NoError(t, err)

	// Run InitCursorRules without force
	err = parser.InitCursorRules(ctx, repo, false)
	require.NoError(t, err)

	// Read the file back and verify it wasn't overwritten
	contents, err := repo.Get(ctx, ".cursor/rules/code-style.mdc")
	require.NoError(t, err)
	require.Equal(t, customContent, contents)
}

func TestInitCursorRulesOverwritesWithForce(t *testing.T) {
	tmp := t.TempDir()

	repo, instanceID, err := cmdutil.RepoForProjectPath(tmp)
	require.NoError(t, err)

	ctx := context.Background()
	err = parser.InitEmpty(ctx, repo, instanceID, "Test Project", "duckdb")
	require.NoError(t, err)

	// Create a cursor rule file with custom content using filesystem
	customContent := "# My Custom Rules\n"
	cursorDir := filepath.Join(tmp, ".cursor", "rules")
	err = os.MkdirAll(cursorDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(cursorDir, "code-style.mdc"), []byte(customContent), 0o644)
	require.NoError(t, err)

	// Run InitCursorRules with force
	err = parser.InitCursorRules(ctx, repo, true)
	require.NoError(t, err)

	// Read the file back and verify it was overwritten
	contents, err := repo.Get(ctx, ".cursor/rules/code-style.mdc")
	require.NoError(t, err)
	require.Contains(t, contents, "# Rill Code Style")
	require.NotContains(t, contents, customContent)
}
