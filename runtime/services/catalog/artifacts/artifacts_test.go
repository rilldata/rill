package artifacts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSourceReadWrite(t *testing.T) {
	catalogs := []struct {
		Name    string
		Catalog *drivers.CatalogObject
	}{
		{
			"Source",
			&drivers.CatalogObject{
				Name: "Source",
				Type: drivers.CatalogObjectTypeSource,
				Path: "path/Source.yaml",
				Definition: toProto(&api.Source{
					Name:      "Source",
					Connector: "file",
					Properties: toProtoStruct(map[string]any{
						"path": "data/source.csv",
					}),
				}),
			},
		},
		{
			"Model",
			&drivers.CatalogObject{
				Name: "Model",
				Type: drivers.CatalogObjectTypeModel,
				Path: "path/Model.yaml",
				Definition: toProto(&api.Model{
					Name:    "Model",
					Sql:     "select * from A",
					Dialect: 0,
				}),
			},
		},
		{
			"MetricsView",
			&drivers.CatalogObject{
				Name: "MetricsView",
				Type: drivers.CatalogObjectTypeMetricsView,
				Path: "path/MetricsView.yaml",
				Definition: toProto(&api.MetricsView{
					Name:          "MetricsView",
					From:          "Model",
					TimeDimension: "time",
					TimeGrains:    []string{"1 day", "1 month"},
					Dimensions: []*api.MetricsView_Dimension{
						{
							Name:        "dim0",
							Type:        "VARCHAR",
							Label:       "Dim0_L",
							Description: "Dim0_D",
							Format:      "humanise",
						},
						{
							Name:        "dim1",
							Type:        "VARCHAR",
							Label:       "Dim1_L",
							Description: "Dim1_D",
							Format:      "humanise",
						},
					},
					Measures: []*api.MetricsView_Measure{
						{
							Name:        "mea0",
							Type:        "INTEGER",
							Label:       "Mea0_L",
							Expression:  "count(c0)",
							Description: "Mea0_D",
							Format:      "humanise",
						},
						{
							Name:        "mea1",
							Type:        "INTEGER",
							Label:       "Mea1_L",
							Expression:  "avg(c1)",
							Description: "Mea1_D",
							Format:      "humanise",
						},
					},
				}),
			},
		},
	}

	fileStore, err := drivers.Open("file", t.TempDir())
	require.NoError(t, err)
	repoStore, _ := fileStore.RepoStore()
	ctx := context.Background()

	for _, tt := range catalogs {
		t.Run(fmt.Sprintf("%s", tt.Name), func(t *testing.T) {
			repo := &api.Repo{
				RepoId: "foo",
				Driver: "file",
				Dsn:    t.TempDir(),
			}
			err := artifacts.Write(ctx, repoStore, repo, tt.Catalog)
			require.NoError(t, err)

			readCatalog, err := artifacts.Read(ctx, repoStore, repo, tt.Catalog.Path)
			require.NoError(t, err)
			require.Equal(t, readCatalog.Name, tt.Catalog.Name)
			require.Equal(t, readCatalog.Path, tt.Catalog.Path)
			require.Equal(t, readCatalog.SQL, tt.Catalog.SQL)
			require.Equal(
				t,
				fromProto(readCatalog.Definition, tt.Catalog.Type),
				fromProto(tt.Catalog.Definition, tt.Catalog.Type),
			)
		})
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

func fromProto(bytes []byte, catalogType string) proto.Message {
	switch catalogType {
	case drivers.CatalogObjectTypeSource:
		var source api.Source
		err := proto.Unmarshal(bytes, &source)
		if err != nil {
			panic(err)
		}
		return &source
	case drivers.CatalogObjectTypeModel:
		var model api.Model
		err := proto.Unmarshal(bytes, &model)
		if err != nil {
			panic(err)
		}
		return &model
	case drivers.CatalogObjectTypeMetricsView:
		var metricsView api.MetricsView
		err := proto.Unmarshal(bytes, &metricsView)
		if err != nil {
			panic(err)
		}
		return &metricsView
	}

	return nil
}
