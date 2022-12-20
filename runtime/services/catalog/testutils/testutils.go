package testutils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func CreateSource(t *testing.T, s *catalog.Service, name string, file string, path string) string {
	absFile, err := filepath.Abs(file)
	require.NoError(t, err)

	ctx := context.Background()
	time.Sleep(time.Millisecond * 10)
	err = artifacts.Write(ctx, s.Repo, s.InstId, &drivers.CatalogEntry{
		Name: name,
		Type: drivers.ObjectTypeSource,
		Path: path,
		Object: &runtimev1.Source{
			Name:      name,
			Connector: "local_file",
			Properties: ToProtoStruct(map[string]any{
				"path": absFile,
			}),
		},
	})
	require.NoError(t, err)
	blob, err := s.Repo.Get(ctx, s.InstId, path)
	require.NoError(t, err)
	return blob
}

func CreateModel(t *testing.T, s *catalog.Service, name string, sql string, path string) string {
	ctx := context.Background()
	time.Sleep(time.Millisecond * 10)
	err := artifacts.Write(ctx, s.Repo, s.InstId, &drivers.CatalogEntry{
		Name: name,
		Type: drivers.ObjectTypeModel,
		Path: path,
		Object: &runtimev1.Model{
			Name:    name,
			Sql:     sql,
			Dialect: runtimev1.Model_DIALECT_DUCKDB,
		},
	})
	require.NoError(t, err)
	blob, err := s.Repo.Get(ctx, s.InstId, path)
	require.NoError(t, err)
	return blob
}

func CreateMetricsView(t *testing.T, s *catalog.Service, metricsView *runtimev1.MetricsView, path string) string {
	ctx := context.Background()
	time.Sleep(time.Millisecond * 10)
	err := artifacts.Write(ctx, s.Repo, s.InstId, &drivers.CatalogEntry{
		Name:   metricsView.Name,
		Type:   drivers.ObjectTypeMetricsView,
		Path:   path,
		Object: metricsView,
	})
	require.NoError(t, err)
	blob, err := s.Repo.Get(ctx, s.InstId, path)
	require.NoError(t, err)
	return blob
}

func ToProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}

func AssertTable(t *testing.T, s *catalog.Service, name string, path string) {
	catalogEntry := AssertInCatalogStore(t, s, name, path)

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

	var schema *runtimev1.StructType
	switch catalogEntry.Type {
	case drivers.ObjectTypeTable:
		schema = catalogEntry.GetTable().Schema
	case drivers.ObjectTypeSource:
		schema = catalogEntry.GetSource().Schema
	case drivers.ObjectTypeModel:
		schema = catalogEntry.GetModel().Schema
	}

	table, err := s.Olap.InformationSchema().Lookup(context.Background(), name)
	require.NoError(t, err)
	require.Equal(t, name, table.Name)
	require.Equal(t, schema.Fields, table.Schema.Fields)
}

func AssertInCatalogStore(t *testing.T, s *catalog.Service, name string, path string) *drivers.CatalogEntry {
	catalogEntry, ok := s.Catalog.FindEntry(context.Background(), s.InstId, name)
	require.True(t, ok)
	require.Equal(t, name, catalogEntry.Name)
	require.Equal(t, path, catalogEntry.Path)
	return catalogEntry
}

func AssertTableAbsence(t *testing.T, s *catalog.Service, name string) {
	_, ok := s.Catalog.FindEntry(context.Background(), s.InstId, name)
	require.False(t, ok)

	_, err := s.Olap.InformationSchema().Lookup(context.Background(), name)
	require.ErrorIs(t, err, drivers.ErrNotFound)
}

func AssertMigration(
	t *testing.T,
	result *catalog.ReconcileResult,
	errCount int,
	addCount int,
	updateCount int,
	dropCount int,
	affectedPaths []string,
) {
	require.Len(t, result.Errors, errCount)
	require.Len(t, result.AddedObjects, addCount)
	require.Len(t, result.UpdatedObjects, updateCount)
	require.Len(t, result.DroppedObjects, dropCount)
	require.ElementsMatch(t, result.AffectedPaths, affectedPaths)
}

func RenameFile(t *testing.T, dir string, from string, to string) {
	time.Sleep(time.Millisecond * 10)
	err := os.Rename(path.Join(dir, from), path.Join(dir, to))
	require.NoError(t, err)
	err = os.Chtimes(path.Join(dir, to), time.Now(), time.Now())
	require.NoError(t, err)
}

func CopyFileToData(t *testing.T, dir string, source string) {
	baseName := path.Base(source)
	dest := path.Join(dir, "data", baseName)

	err := os.MkdirAll(path.Join(dir, "data"), 0777)
	require.NoError(t, err)

	sourceFile, err := os.Open(source)
	require.NoError(t, err)
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	require.NoError(t, err)
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	require.NoError(t, err)
}
