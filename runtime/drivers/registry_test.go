package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRegistry(t *testing.T, reg drivers.RegistryStore) {
	testRegistryInstances(t, reg)
	testRegistryRepos(t, reg)
}

func testRegistryInstances(t *testing.T, reg drivers.RegistryStore) {
	ctx := context.Background()
	inst := &drivers.Instance{
		Driver:       "duckdb",
		DSN:          ":memory:",
		ObjectPrefix: "hello_",
		Exposed:      true,
		EmbedCatalog: true,
	}

	err := reg.CreateInstance(ctx, inst)
	require.NoError(t, err)
	_, err = uuid.Parse(inst.ID)
	require.NoError(t, err)
	require.Equal(t, "duckdb", inst.Driver)
	require.Equal(t, ":memory:", inst.DSN)
	require.Equal(t, "hello_", inst.ObjectPrefix)
	require.Equal(t, true, inst.Exposed)
	require.Equal(t, true, inst.EmbedCatalog)
	require.Greater(t, time.Minute, time.Since(inst.CreatedOn))
	require.Greater(t, time.Minute, time.Since(inst.UpdatedOn))

	res, found := reg.FindInstance(ctx, inst.ID)
	require.True(t, found)
	require.Equal(t, inst.Driver, res.Driver)
	require.Equal(t, inst.DSN, res.DSN)
	require.Equal(t, inst.ObjectPrefix, res.ObjectPrefix)
	require.Equal(t, inst.Exposed, res.Exposed)
	require.Equal(t, inst.EmbedCatalog, res.EmbedCatalog)

	err = reg.CreateInstance(ctx, &drivers.Instance{Driver: "druid"})
	require.NoError(t, err)

	insts := reg.FindInstances(ctx)
	require.Equal(t, 2, len(insts))

	err = reg.DeleteInstance(ctx, inst.ID)
	require.NoError(t, err)

	_, found = reg.FindInstance(ctx, inst.ID)
	require.False(t, found)

	insts = reg.FindInstances(ctx)
	require.Equal(t, 1, len(insts))
}

func testRegistryRepos(t *testing.T, reg drivers.RegistryStore) {
	ctx := context.Background()
	rep := &drivers.Repo{
		Driver: "file",
		DSN:    ".",
	}

	err := reg.CreateRepo(ctx, rep)
	require.NoError(t, err)
	_, err = uuid.Parse(rep.ID)
	require.NoError(t, err)
	require.Equal(t, "file", rep.Driver)
	require.Equal(t, ".", rep.DSN)
	require.Greater(t, time.Minute, time.Since(rep.CreatedOn))
	require.Greater(t, time.Minute, time.Since(rep.UpdatedOn))

	res, found := reg.FindRepo(ctx, rep.ID)
	require.True(t, found)
	require.Equal(t, rep.Driver, res.Driver)
	require.Equal(t, rep.DSN, res.DSN)

	err = reg.CreateRepo(ctx, &drivers.Repo{Driver: "postgres"})
	require.NoError(t, err)

	reps := reg.FindRepos(ctx)
	require.Equal(t, 2, len(reps))

	err = reg.DeleteRepo(ctx, rep.ID)
	require.NoError(t, err)

	_, found = reg.FindRepo(ctx, rep.ID)
	require.False(t, found)

	reps = reg.FindRepos(ctx)
	require.Equal(t, 1, len(reps))
}
