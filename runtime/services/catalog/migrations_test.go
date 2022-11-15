package catalog

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/metrics_views"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

const testDataPath = "../../../web-local/test/data/"

const AdBidsRepoPath = "/sources/AdBids.yaml"
const AdBidsNewRepoPath = "/sources/AdBidsNew.yaml"
const AdBidsModelRepoPath = "/models/AdBids_model.yaml"

func TestMigrate(t *testing.T) {
	configs := []struct {
		title  string
		config MigrationConfig
	}{
		{"MigrateAll", MigrationConfig{}},
		{"MigrateSelected", MigrationConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// same name different content
			createSource(t, s, "AdBids", "AdImpressions.tsv", AdBidsRepoPath)
			result, err := s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 2, 0, 2, 0)
			require.Equal(t, metrics_views.SourceNotFound.Error(), result.Errors[1].Message)
			assertTable(t, s, "AdBids", AdBidsRepoPath)
			assertTableAbsence(t, s, "AdBids_model")

			// revert to stable state
			createSource(t, s, "AdBids", "AdBids.csv", AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			// TODO: should the model/dashboard be counted as updated or added
			assertMigration(t, result, 0, 2, 1, 0)
			assertTable(t, s, "AdBids", AdBidsRepoPath)
			assertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// update with same content
			createSource(t, s, "AdBids", "AdBids.csv", AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 0, 0, 0, 0)

			// delete from olap
			res, err := s.Olap.Execute(context.Background(), &drivers.Statement{
				Query: "drop table AdBids",
			})
			require.NoError(t, err)
			require.NoError(t, res.Close())
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 0, 1, 2, 0)

			// delete file
			err = os.Remove(path.Join(dir, AdBidsRepoPath))
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 2, 0, 1, 1)
			assertTableAbsence(t, s, "AdBids")
			assertTableAbsence(t, s, "AdBids_model")
		})
	}
}

func TestMigrateRenames(t *testing.T) {
	configs := []struct {
		title  string
		config MigrationConfig
	}{
		{"MigrateAll", MigrationConfig{}},
		{"MigrateSelected", MigrationConfig{
			ChangedPaths: []string{AdBidsRepoPath, AdBidsNewRepoPath},
		}},
		{"MigrateSelectedReverseOrder", MigrationConfig{
			ChangedPaths: []string{AdBidsNewRepoPath, AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// write to a new file (should rename)
			err := os.Remove(path.Join(dir, AdBidsRepoPath))
			require.NoError(t, err)
			createSource(t, s, "AdBidsNew", "AdBids.csv", AdBidsNewRepoPath)
			createModel(t, s, "AdBids_model", "select * from AdBidsNew", AdBidsModelRepoPath)
			result, err := s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 0, 0, 3, 0)
			assertTableAbsence(t, s, "AdBids")
			assertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			assertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// write a new file with same name
			createSource(t, s, "AdBidsNew", "AdImpressions.csv", AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 1, 0, 0, 0)
			assertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			assertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestRefreshSource(t *testing.T) {
	configs := []struct {
		title  string
		config MigrationConfig
	}{
		{"MigrateAll", MigrationConfig{
			ForcedPaths: []string{AdBidsRepoPath},
		}},
		{"MigrateSelected", MigrationConfig{
			ForcedPaths:  []string{AdBidsRepoPath},
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			// update with same content
			createSource(t, s, "AdBids", "AdBids.csv", AdBidsRepoPath)
			result, err := s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			// ForcedPaths updates all dependant items
			assertMigration(t, result, 0, 0, 3, 0)
		})
	}
}

func TestInterdependentModel(t *testing.T) {
	configs := []struct {
		title  string
		config MigrationConfig
	}{
		{"MigrateAll", MigrationConfig{}},
		{"MigrateSelected", MigrationConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			AdBidsSourceModelRepoPath := "/models/AdBids_source_model.yaml"

			createModel(t, s, "AdBids_source_model", "select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsSourceModelRepoPath)
			createModel(t, s, "AdBids_model", "select id, timestamp, publisher, domain, bid_price from AdBids_source_model", AdBidsModelRepoPath)
			result, err := s.Migrate(context.Background(), MigrationConfig{})
			require.NoError(t, err)
			assertMigration(t, result, 0, 1, 2, 0)
			assertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			assertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// trigger error in source
			createSource(t, s, "AdBids", "AdImpressions.tsv", AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 3, 0, 3, 0)
			require.Equal(t, metrics_views.SourceNotFound.Error(), result.Errors[2].Message)
			assertTableAbsence(t, s, "AdBids_source_model")
			assertTableAbsence(t, s, "AdBids_model")

			// reset the source
			createSource(t, s, "AdBids", "AdBids.csv", AdBidsRepoPath)
			result, err = s.Migrate(context.Background(), tt.config)
			require.NoError(t, err)
			assertMigration(t, result, 0, 3, 1, 0)
			assertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			assertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestMigrateMetricsView(t *testing.T) {
	s, _ := initBasicService(t)

	createModel(t, s, "AdBids_model", "select id, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err := s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 1, 0, 1, 0)
	// dropping the timestamp column gives a different error
	require.Equal(t, metrics_views.TimestampNotFound.Error(), result.Errors[0].Message)

	createModel(t, s, "AdBids_model", "select id, timestamp, publisher from AdBids", AdBidsModelRepoPath)
	result, err = s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	// invalid measure/dimension doesnt return error for the object
	assertMigration(t, result, 0, 1, 1, 0)
	require.Empty(t, result.AddedObjects[0].MetricsView.Measures[0].Error)
	require.Contains(t, result.AddedObjects[0].MetricsView.Measures[1].Error, `Binder Error: Referenced column "bid_price" not found`)
	require.Empty(t, "", result.AddedObjects[0].MetricsView.Dimensions[0].Error)
	require.Equal(t, result.AddedObjects[0].MetricsView.Dimensions[1].Error, `dimension not found: domain`)
}

func initBasicService(t *testing.T) (*Service, string) {
	s, dir := getService(t)
	createSource(t, s, "AdBids", "AdBids.csv", AdBidsRepoPath)
	result, err := s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 0, 1, 0, 0)
	assertTable(t, s, "AdBids", AdBidsRepoPath)

	createModel(t, s, "AdBids_model", "select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err = s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 0, 1, 0, 0)
	assertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

	createMetricsView(t, s, &api.MetricsView{
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
	result, err = s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 0, 1, 0, 0)
	assertInCatalogStore(t, s, "AdBids_dashboard", "/dashboards/AdBids_dashboard.yaml")

	return s, dir
}

func createSource(t *testing.T, s *Service, name string, file string, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name: name,
		Type: api.CatalogObject_TYPE_SOURCE,
		Source: &api.Source{
			Name:      name,
			Connector: "file",
			Properties: toProtoStruct(map[string]any{
				"path": testDataPath + file,
			}),
		},
		Path: path,
	})
	require.NoError(t, err)
}

func createModel(t *testing.T, s *Service, name string, sql string, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name: name,
		Type: api.CatalogObject_TYPE_MODEL,
		Model: &api.Model{
			Name:    name,
			Sql:     sql,
			Dialect: api.Model_DIALECT_DUCKDB,
		},
		Path: path,
	})
	require.NoError(t, err)
}

func createMetricsView(t *testing.T, s *Service, metricsView *api.MetricsView, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name:        metricsView.Name,
		Type:        api.CatalogObject_TYPE_METRICS_VIEW,
		MetricsView: metricsView,
		Path:        path,
	})
	require.NoError(t, err)
}

func getService(t *testing.T) (*Service, string) {
	duckdbStore, err := drivers.Open("duckdb", "")
	require.NoError(t, err)
	err = duckdbStore.Migrate(context.Background())
	require.NoError(t, err)
	olap, ok := duckdbStore.OLAPStore()
	require.True(t, ok)
	catalog, ok := duckdbStore.CatalogStore()
	require.True(t, ok)

	dir := t.TempDir()
	fileStore, err := drivers.Open("file", dir)
	require.NoError(t, err)
	repo, ok := fileStore.RepoStore()
	require.True(t, ok)

	return NewService(catalog, repo, olap, "test", "test"), dir
}

func toProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}

func assertMigration(t *testing.T, result MigrationResult, errCount int, addCount int, updateCount int, dropCount int) {
	require.Len(t, result.Errors, errCount)
	require.Len(t, result.AddedObjects, addCount)
	require.Len(t, result.UpdatedObjects, updateCount)
	require.Len(t, result.DroppedObjects, dropCount)
}

func assertTable(t *testing.T, s *Service, name string, path string) {
	assertInCatalogStore(t, s, name, path)

	rows, err := s.Olap.Execute(context.Background(), &drivers.Statement{
		Query:    fmt.Sprintf("select count(*) as count from %s", name),
		Args:     nil,
		DryRun:   false,
		Priority: 0,
	})
	require.NoError(t, err)
	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Greater(t, count, 1)
	require.NoError(t, rows.Close())

	table, err := s.Olap.InformationSchema().Lookup(context.Background(), name)
	require.NoError(t, err)
	require.Equal(t, name, table.Name)
}

func assertInCatalogStore(t *testing.T, s *Service, name string, path string) {
	catalog, ok := s.Catalog.FindObject(context.Background(), s.InstId, name)
	require.True(t, ok)
	require.Equal(t, name, catalog.Name)
	require.Equal(t, path, catalog.Path)
}

func assertTableAbsence(t *testing.T, s *Service, name string) {
	_, ok := s.Catalog.FindObject(context.Background(), s.InstId, name)
	require.False(t, ok)

	_, err := s.Olap.InformationSchema().Lookup(context.Background(), name)
	require.ErrorIs(t, err, drivers.ErrNotFound)
}
