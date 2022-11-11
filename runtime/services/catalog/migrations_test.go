package catalog

import (
	"context"
	"fmt"
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

func TestService_MigrateAll(t *testing.T) {
	s := initBasicService(t)

	createSource(t, s, &api.Source{
		Name:      "AdBids",
		Connector: "file",
		Properties: toProtoStruct(map[string]any{
			"path": testDataPath + "AdImpressions.tsv",
		}),
	}, AdBidsRepoPath)
	result, err := s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 2, 0, 2, 0)
	require.ErrorIs(t, result.ArtifactErrors[1].Error, metrics_views.SourceNotFound)
	assertTable(t, s, "AdBids", AdBidsRepoPath)
	assertTableAbsence(t, s, "AdBids_model")

	createSource(t, s, &api.Source{
		Name:      "AdBids",
		Connector: "file",
		Properties: toProtoStruct(map[string]any{
			"path": testDataPath + "AdBids.csv",
		}),
	}, AdBidsRepoPath)
	result, err = s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	// TODO: should the model/dashboard be counted as updated or added
	assertMigration(t, result, 0, 2, 1, 0)
	assertTable(t, s, "AdBids", AdBidsRepoPath)
	assertTable(t, s, "AdBids_model", "/models/AdBids_model.yaml")
}

func TestService_MigrateSelected(t *testing.T) {
	s := initBasicService(t)

	createSource(t, s, &api.Source{
		Name:      "AdBids",
		Connector: "file",
		Properties: toProtoStruct(map[string]any{
			"path": testDataPath + "AdImpressions.tsv",
		}),
	}, AdBidsRepoPath)
	result, err := s.Migrate(context.Background(), MigrationConfig{
		ChangedPaths: []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	assertMigration(t, result, 2, 0, 2, 0)
	require.ErrorIs(t, result.ArtifactErrors[1].Error, metrics_views.SourceNotFound)
	assertTable(t, s, "AdBids", AdBidsRepoPath)
	assertTableAbsence(t, s, "AdBids_model")

	createSource(t, s, &api.Source{
		Name:      "AdBids",
		Connector: "file",
		Properties: toProtoStruct(map[string]any{
			"path": testDataPath + "AdBids.csv",
		}),
	}, AdBidsRepoPath)
	result, err = s.Migrate(context.Background(), MigrationConfig{
		ChangedPaths: []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	// TODO: should the model/dashboard be counted as updated or added
	assertMigration(t, result, 0, 2, 1, 0)
	assertTable(t, s, "AdBids", AdBidsRepoPath)
	assertTable(t, s, "AdBids_model", "/models/AdBids_model.yaml")
}

func TestService_MigrateMetricsView(t *testing.T) {
	s := initBasicService(t)

	createModel(t, s, &api.Model{
		Name:    "AdBids_model",
		Sql:     "select id, publisher, domain, bid_price from AdBids",
		Dialect: api.Model_DuckDB,
	}, "/models/AdBids_model.yaml")
	result, err := s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 1, 0, 1, 0)
	// dropping the timestamp column gives a different error
	require.ErrorIs(t, result.ArtifactErrors[0].Error, metrics_views.TimestampNotFound)

	createModel(t, s, &api.Model{
		Name:    "AdBids_model",
		Sql:     "select id, timestamp, publisher from AdBids",
		Dialect: api.Model_DuckDB,
	}, "/models/AdBids_model.yaml")
	result, err = s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	// invalid measure/dimension doesnt return error for the object
	assertMigration(t, result, 0, 1, 1, 0)
	require.Empty(t, result.AddedObjects[0].MetricsView.Measures[0].Error)
	require.Contains(t, result.AddedObjects[0].MetricsView.Measures[1].Error, `Binder Error: Referenced column "bid_price" not found`)
	require.Empty(t, "", result.AddedObjects[0].MetricsView.Dimensions[0].Error)
	require.Equal(t, result.AddedObjects[0].MetricsView.Dimensions[1].Error, `dimension not found: domain`)
}

func initBasicService(t *testing.T) *Service {
	s := getService(t)
	createSource(t, s, &api.Source{
		Name:      "AdBids",
		Connector: "file",
		Properties: toProtoStruct(map[string]any{
			"path": testDataPath + "AdBids.csv",
		}),
	}, AdBidsRepoPath)
	result, err := s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 0, 1, 0, 0)
	assertTable(t, s, "AdBids", AdBidsRepoPath)

	createModel(t, s, &api.Model{
		Name:    "AdBids_model",
		Sql:     "select id, timestamp, publisher, domain, bid_price from AdBids",
		Dialect: api.Model_DuckDB,
	}, "/models/AdBids_model.yaml")
	result, err = s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	assertMigration(t, result, 0, 1, 0, 0)
	assertTable(t, s, "AdBids_model", "/models/AdBids_model.yaml")

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

	return s
}

func createSource(t *testing.T, s *Service, source *api.Source, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name:   source.Name,
		Type:   api.CatalogObject_TYPE_SOURCE,
		Source: source,
		Path:   path,
	})
	require.NoError(t, err)
}

func createModel(t *testing.T, s *Service, model *api.Model, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name:  model.Name,
		Type:  api.CatalogObject_TYPE_MODEL,
		Model: model,
		Path:  path,
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

func getService(t *testing.T) *Service {
	duckdbStore, err := drivers.Open("duckdb", "")
	require.NoError(t, err)
	err = duckdbStore.Migrate(context.Background())
	require.NoError(t, err)
	olap, ok := duckdbStore.OLAPStore()
	require.True(t, ok)
	catalog, ok := duckdbStore.CatalogStore()
	require.True(t, ok)

	fileStore, err := drivers.Open("file", t.TempDir())
	require.NoError(t, err)
	repo, ok := fileStore.RepoStore()
	require.True(t, ok)

	return NewService(catalog, repo, olap, "test", "test")
}

func toProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}

func assertMigration(t *testing.T, result MigrationResult, errCount int, addCount int, updateCount int, dropCount int) {
	require.Len(t, result.ArtifactErrors, errCount)
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
