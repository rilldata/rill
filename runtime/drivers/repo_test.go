package drivers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRepo(t *testing.T, repo drivers.RepoStore) {
	ctx := context.Background()
	repoID := uuid.NewString()

	paths, err := repo.ListRecursive(ctx, repoID, "**")
	require.NoError(t, err)
	require.Len(t, paths, 0)

	err = repo.PutBlob(ctx, repoID, "foo.sql", "hello world")
	require.NoError(t, err)
	err = repo.PutBlob(ctx, repoID, "/nested/bar.sql", "hello world")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, repoID, "/**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/nested/bar.sql"}, paths)

	paths, err = repo.ListRecursive(ctx, repoID, "./**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/nested/bar.sql"}, paths)

	paths, err = repo.ListRecursive(ctx, repoID, "/nested/**")
	require.NoError(t, err)
	require.Equal(t, []string{"/nested/bar.sql"}, paths)

	err = repo.Delete(ctx, repoID, "nested/bar.sql")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, repoID, "**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql"}, paths)

	_, err = repo.Get(ctx, repoID, "nested/bar.sql")
	require.Error(t, err)

	blob, err := repo.Get(ctx, repoID, "foo.sql")
	require.NoError(t, err)
	require.Equal(t, "hello world", blob)

	err = repo.PutBlob(ctx, repoID, "foo.sql", "bar bar bar")
	require.NoError(t, err)

	blob, err = repo.Get(ctx, repoID, "foo.sql")
	require.NoError(t, err)
	require.Equal(t, "bar bar bar", blob)

	paths, err = repo.ListRecursive(ctx, repoID, "**")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql"}, paths)

	err = repo.PutBlob(ctx, repoID, "foo.yml", "foo foo")
	require.NoError(t, err)
	err = repo.PutBlob(ctx, repoID, "foo.csv", "foo foo")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, repoID, "**/*.{sql,yaml,yml}")
	require.NoError(t, err)
	require.Equal(t, []string{"/foo.sql", "/foo.yml"}, paths)
}
