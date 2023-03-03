package sources_test

import (
	"context"
	"path/filepath"
	"testing"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

const TestDataPath = "../../../../../web-local/test/data"

var AdBidsCsvPath = filepath.Join(TestDataPath, "AdBids.csv")

const AdBidsRepoPath = "/sources/AdBids.yaml"

func TestSourceMigrator_Update(t *testing.T) {
	s, _ := testutils.GetService(t)
	testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{
		SafeSourceRefresh: true,
	})
	require.NoError(t, err)
	require.Len(t, result.Errors, 0)
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)

	// point to invalid file and reconcile
	testutils.CreateSource(t, s, "AdBids", "_"+AdBidsCsvPath, AdBidsRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
		SafeSourceRefresh: true,
	})
	require.NoError(t, err)
	require.Len(t, result.Errors, 1)
	// table is persisted
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
}
