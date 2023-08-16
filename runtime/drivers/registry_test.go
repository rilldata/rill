package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRegistry(t *testing.T, reg drivers.RegistryStore) {
	ctx := context.Background()
	inst := &drivers.Instance{
		OLAPDriver:          "olap",
		RepoDriver:          "repo",
		EmbedCatalog:        true,
		IngestionLimitBytes: 102345,
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": "."},
			},
			{
				Type:   "duckdb",
				Name:   "olap",
				Config: map[string]string{"dsn": ":memory:"},
			},
		},
	}

	err := reg.CreateInstance(ctx, inst)
	require.NoError(t, err)
	_, err = uuid.Parse(inst.ID)
	require.NoError(t, err)
	require.Equal(t, "olap", inst.OLAPDriver)
	require.Equal(t, "repo", inst.RepoDriver)
	require.Equal(t, true, inst.EmbedCatalog)
	require.Greater(t, time.Minute, time.Since(inst.CreatedOn))
	require.Greater(t, time.Minute, time.Since(inst.UpdatedOn))

	res, err := reg.FindInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.Equal(t, inst.OLAPDriver, res.OLAPDriver)
	require.Equal(t, inst.RepoDriver, res.RepoDriver)
	require.Equal(t, inst.EmbedCatalog, res.EmbedCatalog)
	require.Equal(t, inst.IngestionLimitBytes, res.IngestionLimitBytes)
	require.ElementsMatch(t, inst.Connectors, res.Connectors)

	err = reg.CreateInstance(ctx, &drivers.Instance{OLAPDriver: "druid"})
	require.NoError(t, err)

	insts, err := reg.FindInstances(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(insts))

	err = reg.DeleteInstance(ctx, inst.ID)
	require.NoError(t, err)

	_, err = reg.FindInstance(ctx, inst.ID)
	require.EqualError(t, err, drivers.ErrNotFound.Error())

	insts, err = reg.FindInstances(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(insts))
}
