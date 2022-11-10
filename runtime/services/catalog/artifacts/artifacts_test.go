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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSourceReadWrite(t *testing.T) {
	catalogs := []struct {
		Name    string
		Catalog *api.CatalogObject
	}{
		{
			"Source",
			&api.CatalogObject{
				Name: "Source",
				Path: "path/Source.yaml",
				Type: &api.CatalogObject_Source{
					Source: &api.Source{
						Name:      "Source",
						Connector: "file",
						Properties: toProtoStruct(map[string]any{
							"path": "data/source.csv",
						}),
					},
				},
			},
		},
		{
			"Model",
			&api.CatalogObject{
				Name: "Model",
				Path: "path/Model.yaml",
				Type: &api.CatalogObject_Model{
					Model: &api.Model{
						Name:    "Model",
						Sql:     "select * from A",
						Dialect: 0,
					},
				},
			},
		},
		{
			"MetricsView",
			&api.CatalogObject{
				Name: "MetricsView",
				Path: "path/MetricsView.yaml",
				Type: &api.CatalogObject_MetricsView{
					MetricsView: &api.MetricsView{
						Name:          "MetricsView",
						From:          "Model",
						TimeDimension: "time",
						TimeGrains:    []string{"1 day", "1 month"},
						Dimensions: []*api.MetricsView_Dimension{
							{
								Name:        "dim0",
								Label:       "Dim0_L",
								Description: "Dim0_D",
							},
							{
								Name:        "dim1",
								Label:       "Dim1_L",
								Description: "Dim1_D",
							},
						},
						Measures: []*api.MetricsView_Measure{
							{
								Name:        "measure_0",
								Label:       "Mea0_L",
								Expression:  "count(c0)",
								Description: "Mea0_D",
								Format:      "humanise",
							},
							{
								Name:        "measure_1",
								Label:       "Mea1_L",
								Expression:  "avg(c1)",
								Description: "Mea1_D",
								Format:      "humanise",
							},
						},
					},
				},
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
			err := artifacts.Write(ctx, repoStore, repo.RepoId, tt.Catalog)
			require.NoError(t, err)

			readCatalog, err := artifacts.Read(ctx, repoStore, repo.RepoId, tt.Catalog.Path)
			require.NoError(t, err)
			require.Equal(t, readCatalog, tt.Catalog)
		})
	}
}

func toProtoStruct(obj map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(obj)
	if err != nil {
		panic(err)
	}
	return s
}
