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

	paths, err := repo.ListRecursive(ctx, "**")
	require.NoError(t, err)
	require.Len(t, paths, 0)

	err = repo.Put(ctx, "foo.sql", strings.NewReader("hello world"))
	require.NoError(t, err)
	err = repo.Put(ctx, "/nested/bar.sql", strings.NewReader("hello world"))
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, "/**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/nested/bar.sql"}, paths)

	paths, err = repo.ListRecursive(ctx, "./**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/nested/bar.sql"}, paths)

	paths, err = repo.ListRecursive(ctx, "/nested/**")
	require.NoError(t, err)
	require.Equal(t, []string{"/nested/bar.sql"}, paths)

	err = repo.Delete(ctx, "nested/bar.sql")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, "**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql"}, paths)

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

	paths, err = repo.ListRecursive(ctx, "**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql"}, paths)

	err = repo.Put(ctx, "foo.yml", strings.NewReader("foo foo"))
	require.NoError(t, err)
	err = repo.Put(ctx, "foo.csv", strings.NewReader("foo foo"))
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/foo.yml"}, paths)

	// renaming to existing throws error
	err = repo.Rename(ctx, "foo.yml", "foo.sql")
	require.ErrorIs(t, err, drivers.ErrFileAlreadyExists)
	paths, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/foo.yml"}, paths)

	// rename to existing with different case
	err = repo.Rename(ctx, "foo.sql", "FOO.sql")
	require.NoError(t, err)
	paths, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/FOO.sql", "/foo.yml"}, paths)

	// valid rename
	err = repo.Rename(ctx, "foo.yml", "foo_new.yml")
	require.NoError(t, err)
	paths, err = repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/FOO.sql", "/foo_new.yml"}, paths)
}
