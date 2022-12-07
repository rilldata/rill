package drivers_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRepo(t *testing.T, repo drivers.RepoStore) {
	ctx := context.Background()
	instID := uuid.NewString()

	paths, err := repo.ListRecursive(ctx, instID, "**")
	require.NoError(t, err)
	require.Len(t, paths, 0)

	err = repo.Put(ctx, instID, "foo.sql", strings.NewReader("hello world"))
	require.NoError(t, err)
	err = repo.Put(ctx, instID, "/nested/bar.sql", strings.NewReader("hello world"))
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, instID, "/**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/nested/bar.sql"}, paths)

	paths, err = repo.ListRecursive(ctx, instID, "./**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/nested/bar.sql"}, paths)

	paths, err = repo.ListRecursive(ctx, instID, "/nested/**")
	require.NoError(t, err)
	require.Equal(t, []string{"/nested/bar.sql"}, paths)

	err = repo.Delete(ctx, instID, "nested/bar.sql")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, instID, "**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql"}, paths)

	_, err = repo.Get(ctx, instID, "nested/bar.sql")
	require.Error(t, err)

	blob, err := repo.Get(ctx, instID, "foo.sql")
	require.NoError(t, err)
	require.Equal(t, "hello world", blob)

	err = repo.Put(ctx, instID, "foo.sql", strings.NewReader("bar bar bar"))
	require.NoError(t, err)

	blob, err = repo.Get(ctx, instID, "foo.sql")
	require.NoError(t, err)
	require.Equal(t, "bar bar bar", blob)

	paths, err = repo.ListRecursive(ctx, instID, "**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql"}, paths)

	err = repo.Put(ctx, instID, "foo.yml", strings.NewReader("foo foo"))
	require.NoError(t, err)
	err = repo.Put(ctx, instID, "foo.csv", strings.NewReader("foo foo"))
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, instID, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/foo.yml"}, paths)

	// renaming to existing throws error
	err = repo.Rename(ctx, instID, "foo.yml", "foo.sql")
	require.ErrorIs(t, err, drivers.ErrFileAlreadyExists)
	paths, err = repo.ListRecursive(ctx, instID, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/foo.yml"}, paths)

	// rename to existing with different case
	err = repo.Rename(ctx, instID, "foo.sql", "FOO.sql")
	require.NoError(t, err)
	paths, err = repo.ListRecursive(ctx, instID, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/FOO.sql", "/foo.yml"}, paths)

	// valid rename
	err = repo.Rename(ctx, instID, "foo.yml", "foo_new.yml")
	require.NoError(t, err)
	paths, err = repo.ListRecursive(ctx, instID, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/FOO.sql", "/foo_new.yml"}, paths)
}
