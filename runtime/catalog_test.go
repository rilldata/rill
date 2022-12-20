package runtime_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/services/catalog"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/metricsviews"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCatalog(t *testing.T) {
	ctx := context.Background()
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	cat, err := rt.Catalog(ctx, instanceID)
	require.NoError(t, err)

	sourcePath := "/sources/ad_bids_source.yaml"
	modelPath := "/models/ad_bids.sql"
	metricsPath := "/dashboards/ad_bids_metrics.yaml"

	testutils.AssertTable(t, cat, "ad_bids_source", sourcePath)
	testutils.AssertTable(t, cat, "ad_bids", modelPath)

	// force update the source
	res, err := cat.Reconcile(ctx, catalog.ReconcileConfig{
		ChangedPaths: []string{sourcePath},
		ForcedPaths:  []string{sourcePath},
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, res, 0, 0, 3, 0, []string{sourcePath, modelPath, metricsPath})
}
