package initialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddCursorRulesCreatesAllFilesWhenNoneExist(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(cwd) }()

	require.NoError(t, os.Chdir(tmp))

	// Ensure .cursor doesn't exist yet
	_, err = os.Stat(filepath.Join(tmp, ".cursor"))
	require.True(t, os.IsNotExist(err))

	// Run addCursorRules
	err = addCursorRules(false, tmp)
	require.NoError(t, err)

	// Assert directory created
	info, err := os.Stat(filepath.Join(tmp, ".cursor", "rules"))
	require.NoError(t, err)
	require.True(t, info.IsDir())

	// Both template files should be created and contain expected headers
	files := map[string]string{
		"code-style.mdc":        "# Rill Code Style",
		"project-structure.mdc": "# Rill Project Structure",
	}

	for fileName, sample := range files {
		p := filepath.Join(".cursor", "rules", fileName)
		contents, err := os.ReadFile(p)
		require.NoError(t, err)
		require.Contains(t, string(contents), sample)
	}
}

func TestAddCursorRulesBasePathIsFileReturnsError(t *testing.T) {
	tmp := t.TempDir()

	// Create a file and use it as basePath which should cause MkdirAll to fail
	filePath := filepath.Join(tmp, "not_a_dir")
	require.NoError(t, os.WriteFile(filePath, []byte("i am a file"), 0o644))

	err := addCursorRules(false, filePath)
	require.Error(t, err)
}
