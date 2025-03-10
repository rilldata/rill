package rillv1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnvParsing tests repo without a .env file
func TestEnvParsing(t *testing.T) {
	t.Run("without env file", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			`rill.yaml`: ``,
		})

		parser, err := Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)
		require.Equal(t, len(parser.DotEnv), 0)
	})

	t.Run("with single env file", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			"rill.yaml": ``,
			".env": `
TEST=test
FOO=bar
`,
		})

		parser, err := Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)

		mergedEnv := parser.GetDotEnv()

		require.Equal(t, "test", mergedEnv["TEST"])
		require.Equal(t, "bar", mergedEnv["FOO"])
	})

	t.Run("with multiple env files", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			`rill.yaml`: ``,
			".env": `
ROOT_TEST_VAR=root
`,
			"models/.env": `
MODELS_FOLDER_TEST_VAR=models
`,
		})

		parser, err := Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)

		require.Empty(t, parser.Errors)
		mergedEnv := parser.GetDotEnv()
		require.Equal(t, "root", mergedEnv["ROOT_TEST_VAR"], "root-only variable should be preserved")
		require.Equal(t, "models", mergedEnv["MODELS_FOLDER_TEST_VAR"], "models-only variable should be preserved")
	})

	t.Run("env value merge behavior", func(t *testing.T) {
		ctx := context.Background()
		repo := makeRepo(t, map[string]string{
			`rill.yaml`: ``,
			".env": `
SHARED_VAR=root_value
ROOT_ONLY=root_value
`,
			"models/.env": `
SHARED_VAR=models_value
MODELS_ONLY=models_value
`,
			"models/nested/.env": `
SHARED_VAR=nested_value
NESTED_ONLY=nested_value
`,
		})

		parser, err := Parse(ctx, repo, "", "", "duckdb")
		require.NoError(t, err)
		require.Empty(t, parser.Errors)

		mergedEnv := parser.GetDotEnv()

		// Check that variables from all levels exist
		require.Equal(t, "root_value", mergedEnv["ROOT_ONLY"], "root-only variable should be preserved")
		require.Equal(t, "models_value", mergedEnv["MODELS_ONLY"], "models-only variable should be preserved")
		require.Equal(t, "nested_value", mergedEnv["NESTED_ONLY"], "nested-only variable should be preserved")

		// Check that shared variable takes the value from the deepest .env file
		require.Equal(t, "nested_value", mergedEnv["SHARED_VAR"], "shared variable should take value from deepest .env file")
	})
}

func TestEnvReparse(t *testing.T) {
	ctx := context.Background()

	// Create an empty project
	repo := makeRepo(t, map[string]string{`rill.yaml`: ``, ".env": `ROOT_VAR=root_val`})
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, nil)

	// Check that the initial values are correct
	env := p.GetDotEnv()
	require.Equal(t, "root_val", env["ROOT_VAR"], "root-only variable should be preserved")

	// Update the root .env file
	putRepo(t, repo, map[string]string{`rill.yaml`: ``, ".env": `ROOT_VAR=new_root_val`})

	// Reparse
	diff, err := p.Reparse(ctx, []string{".env"})
	require.NoError(t, err)

	require.Equal(t, &Diff{ModifiedDotEnv: true}, diff)

	env = p.GetDotEnv()
	require.Equal(t, "new_root_val", env["ROOT_VAR"], "root-only variable should be updated")

	// Reparse with no changes
	diff, err = p.Reparse(ctx, []string{})
	require.NoError(t, err)
	require.Equal(t, &Diff{}, diff)

	// Add a new .env in a subfolder
	putRepo(t, repo, map[string]string{"models/.env": `MODELS_VAR=models_val`})

	diff, err = p.Reparse(ctx, []string{".env", "models/.env"})
	require.NoError(t, err)

	require.Equal(t, &Diff{ModifiedDotEnv: true}, diff)

	env = p.GetDotEnv()
	require.Equal(t, "new_root_val", env["ROOT_VAR"], "root-only variable should be preserved")
	require.Equal(t, "models_val", env["MODELS_VAR"], "models-only variable should be added")

	// Update the subfolder .env
	putRepo(t, repo, map[string]string{"models/.env": `MODELS_VAR=new_models_val`})

	diff, err = p.Reparse(ctx, []string{"models/.env"})
	require.NoError(t, err)

	require.Equal(t, &Diff{ModifiedDotEnv: true}, diff)

	env = p.GetDotEnv()
	require.Equal(t, "new_root_val", env["ROOT_VAR"], "root-only variable should be preserved")
	require.Equal(t, "new_models_val", env["MODELS_VAR"], "models-only variable should be updated")

	// Make changes to both files
	putRepo(t, repo, map[string]string{".env": `ROOT_VAR=final_root_val`, "models/.env": `MODELS_VAR=final_models_val`})

	diff, err = p.Reparse(ctx, []string{".env", "models/.env"})
	require.NoError(t, err)

	require.Equal(t, &Diff{ModifiedDotEnv: true}, diff)

	env = p.GetDotEnv()

	require.Equal(t, "final_root_val", env["ROOT_VAR"], "root-only variable should be updated")
	require.Equal(t, "final_models_val", env["MODELS_VAR"], "models-only variable should be updated")

	// Remove the subfolder .env
	putRepo(t, repo, map[string]string{"models/.env": ""})

	diff, err = p.Reparse(ctx, []string{"models/.env"})
	require.NoError(t, err)

	require.Equal(t, &Diff{ModifiedDotEnv: true}, diff)

	env = p.GetDotEnv()
	require.Equal(t, "final_root_val", env["ROOT_VAR"], "root-only variable should be preserved")
	require.NotContains(t, env, "MODELS_VAR", "models-only variable should be removed")

	// Remove the root .env
	putRepo(t, repo, map[string]string{".env": ""})

	diff, err = p.Reparse(ctx, []string{".env"})
	require.NoError(t, err)

	require.Equal(t, &Diff{ModifiedDotEnv: true}, diff)

	env = p.GetDotEnv()
	require.Empty(t, env, "all variables should be removed")

	// Reparse with no changes
	diff, err = p.Reparse(ctx, []string{})
	require.NoError(t, err)
	require.Equal(t, &Diff{}, diff)
}
