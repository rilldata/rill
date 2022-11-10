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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const testDataPath = "../../../web-local/test/data/"

func TestService_Migrate(t *testing.T) {
	s := getService(t)

	source := &api.Source{
		Name:      "AdBids",
		Connector: "file",
		Properties: toProtoStruct(map[string]any{
			"path": testDataPath + "AdBids.csv",
		}),
	}
	path := "/sources/AdBids.yaml"

	createSource(t, s, source, path)

	result, err := s.Migrate(context.Background(), MigrationConfig{})
	require.NoError(t, err)
	require.Len(t, result.ArtifactErrors, 0)
	require.Len(t, result.AddedObjects, 1)

	assertSourceInCatalogObject(t, s, source, path)
}

func createSource(t *testing.T, s *Service, source *api.Source, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name: source.Name,
		Type: &api.CatalogObject_Source{Source: source},
		Path: path,
	})
	require.NoError(t, err)
}

func createModel(t *testing.T, s *Service, model *api.Model, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name: model.Name,
		Type: &api.CatalogObject_Model{Model: model},
		Path: path,
	})
	require.NoError(t, err)
}

func createMetricsView(t *testing.T, s *Service, metricsView *api.MetricsView, path string) {
	err := artifacts.Write(context.Background(), s.Repo, s.RepoId, &api.CatalogObject{
		Name: metricsView.Name,
		Type: &api.CatalogObject_MetricsView{MetricsView: metricsView},
		Path: path,
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

	return &Service{
		Catalog: catalog,
		RepoId:  "test",
		Repo:    repo,
		InstId:  "test",
		Olap:    olap,
	}
}

func toProto(message proto.Message) []byte {
	bytes, err := proto.Marshal(message)
	if err != nil {
		panic(err)
	}
	return bytes
}

func toProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}

func assertSourceInCatalogObject(t *testing.T, s *Service, source *api.Source, path string) {
	catalog, ok := s.Catalog.FindObject(context.Background(), s.InstId, source.Name)
	require.True(t, ok)
	require.Equal(t, catalog.Name, source.Name)
	require.Equal(t, catalog.Path, path)

	rows, err := s.Olap.Execute(context.Background(), &drivers.Statement{
		Query:    fmt.Sprintf("select count(*) as count from %s", source.Name),
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
}
