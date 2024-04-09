package drivers_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRepo(t *testing.T, repo drivers.RepoStore) {
	ctx := context.Background()

	files, err := repo.ListRecursive(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{{"/", true}}, files)

	err = repo.Put(ctx, "foo.sql", strings.NewReader("hello world"))
	require.NoError(t, err)
	err = repo.Put(ctx, "/nested/bar.sql", strings.NewReader("hello world"))
	require.NoError(t, err)

	files, err = repo.ListRecursive(ctx, "/**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
		{"/nested", true},
		{"/nested/bar.sql", false},
	}, files)

	files, err = repo.ListRecursive(ctx, "/**", true)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
		{"/nested/bar.sql", false},
	}, files)

	files, err = repo.ListRecursive(ctx, "./**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
		{"/nested", true},
		{"/nested/bar.sql", false},
	}, files)

	files, err = repo.ListRecursive(ctx, "/nested/**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/nested", true},
		{"/nested/bar.sql", false},
	}, files)

	err = repo.Delete(ctx, "nested/bar.sql")
	require.NoError(t, err)

	files, err = repo.ListRecursive(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
		{"/nested", true},
	}, files)

	// deleting a directory
	err = repo.Delete(ctx, "nested")
	require.NoError(t, err)

	files, err = repo.ListRecursive(ctx, "**", false)
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

	files, err = repo.ListRecursive(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/foo.sql", false},
	}, files)

	err = repo.Put(ctx, "foo.yml", strings.NewReader("foo foo"))
	require.NoError(t, err)
	err = repo.Put(ctx, "foo.csv", strings.NewReader("foo foo"))
	require.NoError(t, err)

	files, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
		{"/foo.yml", false},
	}, files)

	// renaming to existing throws error
	err = repo.Rename(ctx, "foo.yml", "foo.sql")
	require.ErrorIs(t, err, drivers.ErrFileAlreadyExists)
	files, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false},
		{"/foo.yml", false},
	}, files)

	// rename to existing with different case
	err = repo.Rename(ctx, "foo.sql", "FOO.sql")
	require.NoError(t, err)
	files, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/FOO.sql", false},
		{"/foo.yml", false},
	}, files)

	// valid rename
	err = repo.Rename(ctx, "foo.yml", "foo_new.yml")
	require.NoError(t, err)
	files, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/FOO.sql", false},
		{"/foo_new.yml", false},
	}, files)

	// create a new folder
	err = repo.MakeDir(ctx, "new_folder")
	require.NoError(t, err)
	files, err = repo.ListRecursive(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true},
		{"/FOO.sql", false},
		{"/foo.csv", false},
		{"/foo_new.yml", false},
		{"/new_folder", true},
	}, files)
}
