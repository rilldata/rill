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
	ctx := context.Background()
	inst := &drivers.Instance{
		OLAPDriver:          "duckdb",
		OLAPDSN:             ":memory:",
		RepoDriver:          "file",
		RepoDSN:             ".",
		EmbedCatalog:        true,
		IngestionLimitBytes: 102345,
	}

	err := reg.CreateInstance(ctx, inst)
	require.NoError(t, err)
	_, err = uuid.Parse(inst.ID)
	require.NoError(t, err)
	require.Equal(t, "duckdb", inst.OLAPDriver)
	require.Equal(t, ":memory:", inst.OLAPDSN)
	require.Equal(t, "file", inst.RepoDriver)
	require.Equal(t, ".", inst.RepoDSN)
	require.Equal(t, true, inst.EmbedCatalog)
	require.Greater(t, time.Minute, time.Since(inst.CreatedOn))
	require.Greater(t, time.Minute, time.Since(inst.UpdatedOn))

	res, err := reg.FindInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.Equal(t, inst.OLAPDriver, res.OLAPDriver)
	require.Equal(t, inst.OLAPDSN, res.OLAPDSN)
	require.Equal(t, inst.RepoDriver, res.RepoDriver)
	require.Equal(t, inst.RepoDSN, res.RepoDSN)
	require.Equal(t, inst.EmbedCatalog, res.EmbedCatalog)
	require.Equal(t, inst.IngestionLimitBytes, res.IngestionLimitBytes)

	err = reg.CreateInstance(ctx, &drivers.Instance{OLAPDriver: "druid"})
	require.NoError(t, err)

	insts, err := reg.FindInstances(ctx)
	require.Equal(t, 2, len(insts))

	err = reg.DeleteInstance(ctx, inst.ID)
	require.NoError(t, err)

	_, err = reg.FindInstance(ctx, inst.ID)
	require.EqualError(t, err, drivers.ErrNotFound.Error())

	insts, err = reg.FindInstances(ctx)
	require.Equal(t, 1, len(insts))
}
