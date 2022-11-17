package catalog_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/api"
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

func TestMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("migrate: skipping test in short mode")
	}

	configs := []struct {
		title  string
		config catalog.MigrationConfig
	}{
		{"MigrateAll", catalog.MigrationConfig{}},
		{"MigrateSelected", catalog.MigrationConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// same name different content
			testutils.CreateSource(t, s, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err := s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 1, 0)
			require.Equal(t, metrics_views.SourceNotFound.Error(), result.Errors[1].Message)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// revert to stable state
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			// TODO: should the model/dashboard be counted as updated or added
			testutils.AssertMigration(t, result, 0, 2, 1, 0)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// update with same content
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 0, 0)

			// delete from olap
			res, err := s.Olap.Execute(context.Background(), &drivers.Statement{
				Query: "drop table AdBids",
			})
			require.NoError(t, err)
			require.NoError(t, res.Close())
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0)

			// delete file
			err = os.Remove(path.Join(dir, AdBidsRepoPath))
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 0, 1)
			testutils.AssertTableAbsence(t, s, "AdBids")
			testutils.AssertTableAbsence(t, s, "AdBids_model")
		})
	}
}

func TestMigrateRenames(t *testing.T) {
	if testing.Short() {
		t.Skip("migrate: skipping test in short mode")
	}

	configs := []struct {
		title  string
		config catalog.MigrationConfig
	}{
		{"MigrateAll", catalog.MigrationConfig{}},
		{"MigrateSelected", catalog.MigrationConfig{
			ChangedPaths: []string{AdBidsRepoPath, AdBidsNewRepoPath},
		}},
		{"MigrateSelectedReverseOrder", catalog.MigrationConfig{
			ChangedPaths: []string{AdBidsNewRepoPath, AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// write to a new file (should rename)
			err := os.Remove(path.Join(dir, AdBidsRepoPath))
			require.NoError(t, err)
			testutils.CreateSource(t, s, "AdBidsNew", AdBidsCsvPath, AdBidsNewRepoPath)
			testutils.CreateModel(t, s, "AdBids_model", "select * from AdBidsNew", AdBidsModelRepoPath)
			result, err := s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 3, 0)
			testutils.AssertTableAbsence(t, s, "AdBids")
			testutils.AssertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// write a new file with same name
			testutils.CreateSource(t, s, "AdBidsNew", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			// name is derived from file path, so there is no error here and AdBids is added
			testutils.AssertMigration(t, result, 0, 1, 0, 0)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestRefreshSource(t *testing.T) {
	if testing.Short() {
		t.Skip("migrate: skipping test in short mode")
	}

	configs := []struct {
		title  string
		config catalog.MigrationConfig
	}{
		{"MigrateAll", catalog.MigrationConfig{
			ForcedPaths: []string{AdBidsRepoPath},
		}},
		{"MigrateSelected", catalog.MigrationConfig{
			ForcedPaths:  []string{AdBidsRepoPath},
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			// update with same content
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err := s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			// ForcedPaths updates all dependant items
			testutils.AssertMigration(t, result, 0, 0, 3, 0)
		})
	}
}

func TestInterdependentModel(t *testing.T) {
	if testing.Short() {
		t.Skip("migrate: skipping test in short mode")
	}

	configs := []struct {
		title  string
		config catalog.MigrationConfig
	}{
		{"MigrateAll", catalog.MigrationConfig{}},
		{"MigrateSelected", catalog.MigrationConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			testutils.CreateModel(t, s, "AdBids_source_model",
				"select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsSourceModelRepoPath)
			testutils.CreateModel(t, s, "AdBids_model",
				"select id, timestamp, publisher, domain, bid_price from AdBids_source_model", AdBidsModelRepoPath)
			result, err := s.Migrate(context.Background(), catalog.MigrationConfig{})
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0)
			testutils.AssertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// trigger error in source
			testutils.CreateSource(t, s, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 3, 0, 1, 0)
			require.Equal(t, metrics_views.SourceNotFound.Error(), result.Errors[2].Message)
			testutils.AssertTableAbsence(t, s, "AdBids_source_model")
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// reset the source
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 1, 0)
			testutils.AssertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestModelVariations(t *testing.T) {
	if testing.Short() {
		t.Skip("migrate: skipping test in short mode")
	}

	s, _ := initBasicService(t)

	// update to invalid model
	testutils.CreateModel(t, s, "AdBids_model",
		"select id, timestamp, publisher, domain, bid_price AdBids", AdBidsModelRepoPath)
	result, err := s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 0, 0)
	testutils.AssertTableAbsence(t, s, "AdBids_model")

	// new invalid model
	testutils.CreateModel(t, s, "AdBids_source_model",
		"select id, timestamp, publisher, domain, bid_price AdBids", AdBidsSourceModelRepoPath)
	result, err = s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0)
	testutils.AssertTableAbsence(t, s, "AdBids_source_model")
}

func TestMigrateMetricsView(t *testing.T) {
	if testing.Short() {
		t.Skip("migrate: skipping test in short mode")
	}

	s, _ := initBasicService(t)

	testutils.CreateModel(t, s, "AdBids_model", "select id, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err := s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 1, 0)
	// dropping the timestamp column gives a different error
	require.Equal(t, metrics_views.TimestampNotFound.Error(), result.Errors[0].Message)

	testutils.CreateModel(t, s, "AdBids_model", "select id, timestamp, publisher from AdBids", AdBidsModelRepoPath)
	result, err = s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	// invalid measure/dimension doesnt return error for the object
	testutils.AssertMigration(t, result, 0, 1, 1, 0)
	require.Empty(t, result.AddedObjects[0].MetricsView.Measures[0].Error)
	require.Contains(t, result.AddedObjects[0].MetricsView.Measures[1].Error, `Binder Error: Referenced column "bid_price" not found`)
	require.Empty(t, "", result.AddedObjects[0].MetricsView.Dimensions[0].Error)
	require.Equal(t, result.AddedObjects[0].MetricsView.Dimensions[1].Error, `dimension not found: domain`)
}

func initBasicService(t *testing.T) (*catalog.Service, string) {
	s, dir := getService(t)
	testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	result, err := s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0)
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)

	testutils.CreateModel(t, s, "AdBids_model", "select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err = s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0)
	testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

	testutils.CreateMetricsView(t, s, &api.MetricsView{
		Name:          "AdBids_dashboard",
		From:          "AdBids_model",
		TimeDimension: "timestamp",
		TimeGrains:    []string{"1 day", "1 month"},
		Dimensions: []*api.MetricsView_Dimension{
			{
				Name:  "publisher",
				Label: "Publisher",
			},
			{
				Name:  "domain",
				Label: "Domain",
			},
		},
		Measures: []*api.MetricsView_Measure{
			{
				Expression: "count(*)",
			},
			{
				Expression: "avg(bid_price)",
			},
		},
	}, "/dashboards/AdBids_dashboard.yaml")
	result, err = s.Migrate(context.Background(), catalog.MigrationConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0)
	testutils.AssertInCatalogStore(t, s, "AdBids_dashboard", "/dashboards/AdBids_dashboard.yaml")

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

	return catalog.NewService(catalogObject, repo, olap, "test", "test"), dir
}
