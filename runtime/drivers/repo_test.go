package drivers_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRepo(t *testing.T, repo drivers.RepoStore) {
	ctx := context.Background()

	files, err := repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{{"/", true, 0}}, files)

	err = repo.Put(ctx, "foo.sql", strings.NewReader("hello world"))
	require.NoError(t, err)
	err = repo.Put(ctx, "/nested/bar.sql", strings.NewReader("hello world"))
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "/**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true, 0},
		{"/foo.sql", false, 11},
		{"/nested", true, 0},
		{"/nested/bar.sql", false, 11},
	}, files)

	files, err = repo.ListGlob(ctx, "/foo.sql", true)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false, 11},
	}, files)

	files, err = repo.ListGlob(ctx, "/**", true)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false, 11},
		{"/nested/bar.sql", false, 11},
	}, files)

	files, err = repo.ListGlob(ctx, "./**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true, 0},
		{"/foo.sql", false, 11},
		{"/nested", true, 0},
		{"/nested/bar.sql", false, 11},
	}, files)

	files, err = repo.ListGlob(ctx, "/nested/**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/nested", true, 0},
		{"/nested/bar.sql", false, 11},
	}, files)

	err = repo.Delete(ctx, "nested/bar.sql", false)
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true, 0},
		{"/foo.sql", false, 11},
		{"/nested", true, 0},
	}, files)

	// deleting a directory
	err = repo.Delete(ctx, "nested", false)
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true, 0},
		{"/foo.sql", false, 11},
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
		{"/", true, 0},
		{"/foo.sql", false, 11},
	}, files)

	err = repo.Put(ctx, "foo.yml", strings.NewReader("foo foo"))
	require.NoError(t, err)
	err = repo.Put(ctx, "foo.csv", strings.NewReader("foo foo"))
	require.NoError(t, err)

	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false, 11},
		{"/foo.yml", false, 7},
	}, files)

	// renaming to existing throws error
	err = repo.Rename(ctx, "foo.yml", "foo.sql")
	require.ErrorIs(t, err, os.ErrExist)
	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/foo.sql", false, 11},
		{"/foo.yml", false, 7},
	}, files)

	// rename to existing with different case
	err = repo.Rename(ctx, "foo.sql", "FOO.sql")
	require.NoError(t, err)
	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/FOO.sql", false, 11},
		{"/foo.yml", false, 7},
	}, files)

	// valid rename
	err = repo.Rename(ctx, "foo.yml", "foo_new.yml")
	require.NoError(t, err)
	files, err = repo.ListGlob(ctx, "**/*.{sql,yaml,yml}", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/FOO.sql", false, 11},
		{"/foo_new.yml", false, 7},
	}, files)

	// create a new folder
	err = repo.MkdirAll(ctx, "new_folder")
	require.NoError(t, err)
	files, err = repo.ListGlob(ctx, "**", false)
	require.NoError(t, err)
	require.Equal(t, []drivers.DirEntry{
		{"/", true, 0},
		{"/FOO.sql", false, 11},
		{"/foo.csv", false, 7},
		{"/foo_new.yml", false, 7},
		{"/new_folder", true, 0},
	}, files)
}
