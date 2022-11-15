package artifacts_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSourceReadWrite(t *testing.T) {
	catalogs := []struct {
		Name    string
		Catalog *api.CatalogObject
		Raw     string
	}{
		{
			"Source",
			&api.CatalogObject{
				Name: "Source",
				Path: "sources/Source.yaml",
				Type: api.CatalogObject_TYPE_SOURCE,
				Source: &api.Source{
					Name:      "Source",
					Connector: "file",
					Properties: toProtoStruct(map[string]any{
						"path": "data/source.csv",
					}),
				},
			},
			`version: 0.0.1
type: file
uri: data/source.csv
`,
		},
		{
			"Model",
			&api.CatalogObject{
				Name: "Model",
				Path: "models/Model.sql",
				Type: api.CatalogObject_TYPE_MODEL,
				Model: &api.Model{
					Name:    "Model",
					Sql:     "select * from A",
					Dialect: api.Model_DIALECT_DUCKDB,
				},
			},
			"select * from A",
		},
		{
			"MetricsView",
			&api.CatalogObject{
				Name: "MetricsView",
				Path: "dashboards/MetricsView.yaml",
				Type: api.CatalogObject_TYPE_METRICS_VIEW,
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
			`version: 0.0.1
display_name: ""
description: ""
from: Model
time_dimension: time
timegrains:
- 1 day
- 1 month
default_timegrain: ""
dimensions:
- label: Dim0_L
  property: dim0
  description: Dim0_D
- label: Dim1_L
  property: dim1
  description: Dim1_D
measures:
- label: Mea0_L
  expression: count(c0)
  description: Mea0_D
  format_preset: humanise
- label: Mea1_L
  expression: avg(c1)
  description: Mea1_D
  format_preset: humanise
`,
		},
	}

	dir := t.TempDir()
	fileStore, err := drivers.Open("file", dir)
	require.NoError(t, err)
	repoStore, _ := fileStore.RepoStore()
	ctx := context.Background()

	for _, tt := range catalogs {
		t.Run(fmt.Sprintf("%s", tt.Name), func(t *testing.T) {
			repo := &api.Repo{
				RepoId: "foo",
				Driver: "file",
				Dsn:    dir,
			}
			err := artifacts.Write(ctx, repoStore, repo.RepoId, tt.Catalog)
			require.NoError(t, err)

			readCatalog, err := artifacts.Read(ctx, repoStore, repo.RepoId, tt.Catalog.Path)
			require.NoError(t, err)
			require.Equal(t, readCatalog, tt.Catalog)

			b, err := os.ReadFile(path.Join(dir, tt.Catalog.Path))
			require.NoError(t, err)
			require.Equal(t, tt.Raw, string(b))
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
