package drivers_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRepo(t *testing.T, repo drivers.RepoStore) {
	ctx := context.Background()

	files, err := repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{{"/", true}}, files)

	err = repo.Put(ctx, "foo.sql", strings.NewReader("hello world"))
	require.NoError(t, err)
	err = repo.Put(ctx, "/nested/bar.sql", strings.NewReader("hello world"))
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "/**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
		{"/nested", true},
		{"/nested/bar.sql", false},
	}, files)

	files, err = repo.ListGlob(ctx, "/foo.sql", true)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
	}, files)

	files, err = repo.ListGlob(ctx, "/**", true)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
		{"/nested/bar.sql", false},
	}, files)

	files, err = repo.ListGlob(ctx, "./**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
		{"/nested", true},
		{"/nested/bar.sql", false},
	}, files)

	files, err = repo.ListGlob(ctx, "/nested/**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/nested", true},
		{"/nested/bar.sql", false},
	}, files)

	err = repo.Delete(ctx, "nested/bar.sql", false)
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
		{"/nested", true},
	}, files)

	// deleting a directory
	err = repo.Delete(ctx, "nested", false)
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
	}, files)

	_, err = repo.Get(ctx, "nested/bar.sql")
	require.Error(t, err)

	blob, err := repo.Get(ctx, "foo.sql")
	require.NoError(t, err)
	require.Equal(t, "hello world", blob)

	err = repo.Put(ctx, "foo.sql", strings.NewReader("bar bar bar"))
	require.NoError(t, err)

	blob, err = repo.Get(ctx, "foo.sql")
	require.NoError(t, err)
	require.Equal(t, "bar bar bar", blob)

	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
	}, files)

	err = repo.Put(ctx, "foo.yml", strings.NewReader("foo foo"))
	require.NoError(t, err)
	err = repo.Put(ctx, "foo.csv", strings.NewReader("foo foo"))
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
		{"/foo.yml", false},
	}, files)

	// renaming to existing throws error
	err = repo.Rename(ctx, "foo.yml", "foo.sql")
	require.ErrorIs(t, err, os.ErrExist)
	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
		{"/foo.yml", false},
	}, files)

	// rename to existing with different case
	err = repo.Rename(ctx, "foo.sql", "FOO.sql")
	require.NoError(t, err)
	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/FOO.sql", false},
		{"/foo.yml", false},
	}, files)

	// valid rename
	err = repo.Rename(ctx, "foo.yml", "foo_new.yml")
	require.NoError(t, err)
	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/FOO.sql", false},
		{"/foo_new.yml", false},
	}, files)

	// create a new folder
	err = repo.MkdirAll(ctx, "new_folder")
	require.NoError(t, err)
	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/FOO.sql", false},
		{"/foo.csv", false},
		{"/foo_new.yml", false},
		{"/new_folder", true},
	}, files)

	// paths that traverse outside the repo root must be rejected
	_, err = repo.Get(ctx, "/../secret.txt")
	require.ErrorContains(t, err, "outside the repo root")
	_, err = repo.Stat(ctx, "../secret.txt")
	require.ErrorContains(t, err, "outside the repo root")
	err = repo.Put(ctx, "../escape.sql", strings.NewReader("boom"))
	require.ErrorContains(t, err, "outside the repo root")
	err = repo.MkdirAll(ctx, "/nested/../../escape_dir")
	require.ErrorContains(t, err, "outside the repo root")
	err = repo.Rename(ctx, "foo.csv", "../escape.csv")
	require.ErrorContains(t, err, "outside the repo root")
	err = repo.Rename(ctx, "../escape.csv", "foo2.csv")
	require.ErrorContains(t, err, "outside the repo root")
	err = repo.Delete(ctx, "/../escape.sql", false)
	require.ErrorContains(t, err, "outside the repo root")
}

func TestResolveRepoPath(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "data", "repo")

	valid := map[string]string{
		"/models/foo.sql":      filepath.Join(root, "models", "foo.sql"),
		"models/foo.sql":       filepath.Join(root, "models", "foo.sql"),
		"/models/../rill.yaml": filepath.Join(root, "rill.yaml"),
		"/":                    root,
		"":                     root,
	}
	for path, expected := range valid {
		fp, err := drivers.ResolveRepoPath(root, path)
		require.NoError(t, err, "path %q", path)
		require.Equal(t, expected, fp, "path %q", path)
	}

	invalid := []string{
		"..",
		"/..",
		"../secret.txt",
		"/../secret.txt",
		"/models/../../secret.txt",
		"../../../../etc/passwd",
	}
	for _, path := range invalid {
		_, err := drivers.ResolveRepoPath(root, path)
		require.ErrorContains(t, err, "outside the repo root", "path %q", path)
	}
}
