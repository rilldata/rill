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

	paths, err := repo.ListRecursive(ctx, repoID)
	require.NoError(t, err)
	require.Len(t, paths, 0)

	err = repo.Put(ctx, repoID, "foo.sql", "hello world")
	require.NoError(t, err)
	err = repo.Put(ctx, repoID, "/bar.sql", "hello world")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, repoID)
	require.NoError(t, err)
	require.Equal(t, []string{"/bar.sql", "/foo.sql"}, paths)

	err = repo.Put(ctx, repoID, "deeply/nested/foo.sql", "hello world")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, repoID)
	require.NoError(t, err)
	require.Equal(t, []string{"/bar.sql", "/deeply/nested/foo.sql", "/foo.sql"}, paths)

	err = repo.Delete(ctx, repoID, "deeply/nested/foo.sql")
	require.NoError(t, err)

	paths, err = repo.ListRecursive(ctx, repoID)
	require.NoError(t, err)
	require.Equal(t, []string{"/bar.sql", "/foo.sql"}, paths)

	_, err = repo.Get(ctx, repoID, "deeply/nested/foo.sql")
	require.Error(t, err)

	blob, err := repo.Get(ctx, repoID, "bar.sql")
	require.NoError(t, err)
	require.Equal(t, "hello world", blob)

	err = repo.Put(ctx, repoID, "bar.sql", "bar bar bar")
	require.NoError(t, err)

	blob, err = repo.Get(ctx, repoID, "bar.sql")
	require.NoError(t, err)
	require.Equal(t, "bar bar bar", blob)
}
