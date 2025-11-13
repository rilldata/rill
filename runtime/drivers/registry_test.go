package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func testRegistry(t *testing.T, reg drivers.RegistryStore) {
	ctx := context.Background()
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": "."})),
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
			},
			{
				Type:   "sqlite",
				Name:   "catalog",
				Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": "file:rill?mode=memory&cache=shared"})),
			},
		},
	}

	err := reg.CreateInstance(ctx, inst)
	require.NoError(t, err)
	_, err = uuid.Parse(inst.ID)
	require.NoError(t, err)
	require.Equal(t, "test", inst.Environment)
	require.Equal(t, "duckdb", inst.OLAPConnector)
	require.Equal(t, "repo", inst.RepoConnector)
	require.Equal(t, "catalog", inst.CatalogConnector)
	require.Greater(t, time.Minute, time.Since(inst.CreatedOn))
	require.Greater(t, time.Minute, time.Since(inst.UpdatedOn))

	// edit instance
	inst.ProjectDisplayName = "My Project"
	err = reg.EditInstance(ctx, inst)
	require.NoError(t, err)

	res, err := reg.FindInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.Equal(t, inst.OLAPConnector, res.OLAPConnector)
	require.Equal(t, inst.RepoConnector, res.RepoConnector)
	require.Equal(t, inst.CatalogConnector, res.CatalogConnector)
	require.Equal(t, "My Project", res.ProjectDisplayName)
	require.ElementsMatch(t, inst.Connectors, res.Connectors)

	err = reg.CreateInstance(ctx, &drivers.Instance{OLAPConnector: "druid"})
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
