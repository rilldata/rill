package catalog_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/metrics_views"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

const TestDataPath = "../../../web-local/test/data"

var AdBidsCsvPath = filepath.Join(TestDataPath, "AdBids.csv")
var AdImpressionsCsvPath = filepath.Join(TestDataPath, "AdImpressions.tsv")

const AdBidsRepoPath = "/sources/AdBids.yaml"
const AdBidsNewRepoPath = "/sources/AdBidsNew.yaml"
const AdBidsModelRepoPath = "/models/AdBids_model.sql"
const AdBidsSourceModelRepoPath = "/models/AdBids_source_model.sql"
const AdBidsDashboardRepoPath = "/dashboards/AdBids_dashboard.yaml"

var AdBidsAffectedPaths = []string{AdBidsRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
var AdBidsNewAffectedPaths = []string{AdBidsNewRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
var AdBidsDashboardAffectedPaths = []string{AdBidsModelRepoPath, AdBidsDashboardRepoPath}

func TestReconcile(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// same name different content
			testutils.CreateSource(t, s, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 1, 0, AdBidsAffectedPaths)
			require.Equal(t, metrics_views.SourceNotFound, result.Errors[1].Message)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// revert to stable state
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			// TODO: should the model/dashboard be counted as updated or added
			testutils.AssertMigration(t, result, 0, 2, 1, 0, AdBidsAffectedPaths)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// update with same content
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 0, 0, []string{})

			// delete from olap
			res, err := s.Olap.Execute(context.Background(), &drivers.Statement{
				Query: "drop table AdBids",
			})
			require.NoError(t, err)
			require.NoError(t, res.Close())
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0, AdBidsAffectedPaths)

			// delete file
			err = os.Remove(path.Join(dir, AdBidsRepoPath))
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 0, 1, AdBidsAffectedPaths)
			testutils.AssertTableAbsence(t, s, "AdBids")
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// add back source
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 0, 0, AdBidsAffectedPaths)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestReconcileRenames(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath, AdBidsNewRepoPath},
		}},
		{"ReconcileSelectedReverseOrder", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewRepoPath, AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// write to a new file (should rename)
			testutils.RenameFile(t, dir, AdBidsRepoPath, AdBidsNewRepoPath)
			testutils.CreateModel(t, s, "AdBids_model", "select * from AdBidsNew", AdBidsModelRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 3, 0, AdBidsNewAffectedPaths)
			testutils.AssertTableAbsence(t, s, "AdBids")
			testutils.AssertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// write a new file with same name
			testutils.CreateSource(t, s, "AdBidsNew", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			// name is derived from file path, so there is no error here and AdBids is added
			testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsRepoPath})
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestRefreshSource(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{
			ForcedPaths: []string{AdBidsRepoPath},
		}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ForcedPaths:  []string{AdBidsRepoPath},
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			// update with same content
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			// ForcedPaths updates all dependant items
			testutils.AssertMigration(t, result, 0, 0, 3, 0, AdBidsAffectedPaths)
		})
	}
}

func TestInterdependentModel(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	AdBidsSourceAffectedPaths := []string{AdBidsSourceModelRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
	AdBidsAllAffectedPaths := append([]string{AdBidsRepoPath}, AdBidsSourceAffectedPaths...)

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			testutils.CreateModel(t, s, "AdBids_source_model",
				"select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsSourceModelRepoPath)
			testutils.CreateModel(t, s, "AdBids_model",
				"select id, timestamp, publisher, domain, bid_price from AdBids_source_model", AdBidsModelRepoPath)
			result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0, AdBidsSourceAffectedPaths)
			testutils.AssertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// trigger error in source
			testutils.CreateSource(t, s, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 3, 0, 1, 0, AdBidsAllAffectedPaths)
			require.Equal(t, metrics_views.SourceNotFound, result.Errors[2].Message)
			testutils.AssertTableAbsence(t, s, "AdBids_source_model")
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// reset the source
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 1, 0, AdBidsAllAffectedPaths)
			testutils.AssertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestModelRename(t *testing.T) {
	var AdBidsRenameModelRepoPath = "/models/AdBidsRename.sql"
	var AdBidsRenameNewModelRepoPath = "/models/AdBidsRenameNew.sql"

	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRenameModelRepoPath, AdBidsRenameNewModelRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CreateModel(t, s, "AdBidsRename", "select * from AdBids", AdBidsRenameModelRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsRenameModelRepoPath})

			for i := 0; i < 5; i++ {
				testutils.RenameFile(t, dir, AdBidsRenameModelRepoPath, AdBidsRenameNewModelRepoPath)
				result, err = s.Reconcile(context.Background(), tt.config)
				require.NoError(t, err)
				testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsRenameNewModelRepoPath})

				testutils.RenameFile(t, dir, AdBidsRenameNewModelRepoPath, AdBidsRenameModelRepoPath)
				result, err = s.Reconcile(context.Background(), tt.config)
				require.NoError(t, err)
				testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsRenameModelRepoPath})
			}
		})
	}
}

func TestModelVariations(t *testing.T) {
	s, _ := initBasicService(t)

	// update to invalid model
	testutils.CreateModel(t, s, "AdBids_model",
		"select id, timestamp, publisher, domain, bid_price AdBids", AdBidsModelRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 0, 0, AdBidsDashboardAffectedPaths)
	testutils.AssertTableAbsence(t, s, "AdBids_model")

	// new invalid model
	testutils.CreateModel(t, s, "AdBids_source_model",
		"select id, timestamp, publisher, domain, bid_price AdBids", AdBidsSourceModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{AdBidsSourceModelRepoPath})
	testutils.AssertTableAbsence(t, s, "AdBids_source_model")
}

func TestReconcileMetricsView(t *testing.T) {
	s, _ := initBasicService(t)

	testutils.CreateModel(t, s, "AdBids_model", "select id, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 1, 0, AdBidsDashboardAffectedPaths)
	// dropping the timestamp column gives a different error
	require.Equal(t, metrics_views.TimestampNotFound, result.Errors[0].Message)

	testutils.CreateModel(t, s, "AdBids_model", "select id, timestamp, publisher from AdBids", AdBidsModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 1, 0, AdBidsDashboardAffectedPaths)
	require.Equal(t, `dimension not found: domain`, result.Errors[0].Message)
	require.Equal(t, []string{"Dimensions", "1"}, result.Errors[0].PropertyPath)
	require.Contains(t, result.Errors[1].Message, `Binder Error: Referenced column "bid_price" not found`)
	require.Equal(t, []string{"Measures", "1"}, result.Errors[1].PropertyPath)
}

func TestInvalidFiles(t *testing.T) {
	s, _ := initBasicService(t)
	ctx := context.Background()

	err := s.Repo.PutBlob(ctx, s.InstId, AdBidsRepoPath, `version: 0.0.1
type: file
path:
 - data/source.csv`)
	require.NoError(t, err)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 3, 0, 0, 1, AdBidsAffectedPaths)
	require.Contains(t, result.Errors[0].Message, "yaml: unmarshal errors")

	testutils.CreateSource(t, s, "Ad-Bids", "AdBids.csv", "/sources/Ad-Bids.yaml")
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{"/sources/Ad-Bids.yaml"})
	require.Equal(t, "/sources/Ad-Bids.yaml", result.Errors[0].FilePath)
	require.Equal(t, "invalid file name", result.Errors[0].Message)
}

func initBasicService(t *testing.T) (*catalog.Service, string) {
	s, dir := getService(t)
	testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsRepoPath})
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)

	testutils.CreateModel(t, s, "AdBids_model",
		"select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsModelRepoPath})
	testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

	testutils.CreateMetricsView(t, s, &runtimev1.MetricsView{
		Name:          "AdBids_dashboard",
		From:          "AdBids_model",
		TimeDimension: "timestamp",
		TimeGrains:    []string{"1 day", "1 month"},
		Dimensions: []*runtimev1.MetricsView_Dimension{
			{
				Name:  "publisher",
				Label: "Publisher",
			},
			{
				Name:  "domain",
				Label: "Domain",
			},
		},
		Measures: []*runtimev1.MetricsView_Measure{
			{
				Expression: "count(*)",
			},
			{
				Expression: "avg(bid_price)",
			},
		},
	}, AdBidsDashboardRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsDashboardRepoPath})
	testutils.AssertInCatalogStore(t, s, "AdBids_dashboard", AdBidsDashboardRepoPath)

	return s, dir
}

func getService(t *testing.T) (*catalog.Service, string) {
	dir := t.TempDir()

	duckdbStore, err := drivers.Open("duckdb", filepath.Join(dir, "stage.db"))
	require.NoError(t, err)
	err = duckdbStore.Migrate(context.Background())
	require.NoError(t, err)
	olap, ok := duckdbStore.OLAPStore()
	require.True(t, ok)
	catalogObject, ok := duckdbStore.CatalogStore()
	require.True(t, ok)

	fileStore, err := drivers.Open("file", dir)
	require.NoError(t, err)
	repo, ok := fileStore.RepoStore()
	require.True(t, ok)

	return catalog.NewService(catalogObject, repo, olap, "test"), dir
}
