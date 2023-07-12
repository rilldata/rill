package rillv1

import (
	"context"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	_ "github.com/rilldata/rill/runtime/drivers/file"
)

func TestRillYAML(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
title: Hello world
description: This project says hello to the world

connectors:
- name: my-s3
  type: s3
  defaults:
    region: us-east-1

env:
  foo: bar
`,
	})

	res, err := ParseRillYAML(ctx, repo, "")
	require.NoError(t, err)

	require.Equal(t, res.Title, "Hello world")
	require.Equal(t, res.Description, "This project says hello to the world")

	require.Len(t, res.Connectors, 1)
	require.Equal(t, "my-s3", res.Connectors[0].Name)
	require.Equal(t, "s3", res.Connectors[0].Type)
	require.Len(t, res.Connectors[0].Defaults, 1)
	require.Equal(t, "us-east-1", res.Connectors[0].Defaults["region"])

	require.Len(t, res.Variables, 1)
	require.Equal(t, "foo", res.Variables[0].Name)
	require.Equal(t, "bar", res.Variables[0].Default)
}

func TestComplete(t *testing.T) {
	ctx := context.Background()
	truth := true
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// init.sql
		`init.sql`: `
{{ configure "version" 2 }}
INSTALL 'hello';
`,
		// source s1
		`sources/s1.yaml`: `
connector: s3
path: hello
`,
		// source s2
		`sources/s2.sql`: `
{{ configure "connector" "postgres" }}
SELECT 1
`,
		// model m1
		`models/m1.sql`: `
SELECT 1
`,
		// model m2
		`models/m2.sql`: `
SELECT * FROM m1
`,
		`models/m2.yaml`: `
materialize: true
`,
		// dashboard d1
		`dashboards/d1.yaml`: `
model: m2
dimensions:
  - name: a
measures:
  - name: b
    expression: count(*)
`,
		// migration c1
		`custom/c1.yml`: `
kind: migration
version: 3
sql: |
  CREATE TABLE a(a integer);
`,
		// model c2
		`custom/c2.sql`: `
{{ configure "kind" "model" }}
{{ configure "materialize" true }}
SELECT * FROM {{ ref "m2" }}
`,
	}
	repo := makeRepo(t, files)

	expected := []*Resource{
		// init.sql
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "init"},
			Paths: []string{"/init.sql"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Version: 2,
				Sql:     strings.TrimSpace(files["init.sql"]),
			},
		},
		// source s1
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
			Paths: []string{"/sources/s1.yaml"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "s3",
				Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			},
		},
		// source s2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "s2"},
			Paths: []string{"/sources/s2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:      "postgres",
				Sql:            strings.TrimSpace(files["sources/s2.sql"]),
				UsesTemplating: true,
				Materialize:    &truth,
			},
		},
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Sql: strings.TrimSpace(files["models/m1.sql"]),
			},
		},
		// model m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Refs:  []ResourceName{{Name: "m1"}},
			Paths: []string{"/models/m2.yaml", "/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Sql:         strings.TrimSpace(files["models/m2.sql"]),
				Materialize: &truth,
			},
		},
		// dashboard d1
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Refs:  []ResourceName{{Name: "m2"}},
			Paths: []string{"/dashboards/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Model: "m2",
				Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
					{Name: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_Measure{
					{Name: "b", Expression: "count(*)"},
				},
			},
		},
		// migration c1
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "c1"},
			Paths: []string{"/custom/c1.yml"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Version: 3,
				Sql:     "CREATE TABLE a(a integer);",
			},
		},
		// model c2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "c2"},
			Refs:  []ResourceName{{Name: "m2"}},
			Paths: []string{"/custom/c2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Sql:            strings.TrimSpace(files["custom/c2.sql"]),
				Materialize:    &truth,
				UsesTemplating: true,
			},
		},
	}

	p, err := Parse(ctx, repo, "", []string{""})
	require.NoError(t, err)

	for _, want := range expected {
		found := false
		for _, got := range p.Resources {
			if want.Name == got.Name {
				require.Equal(t, want, got)
				delete(p.Resources, got.Name)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing resource %v", want.Name)
		}
	}
	if len(p.Resources) > 0 {
		t.Errorf("unexpected resources: %v", p.Resources)
	}
}

func makeRepo(t *testing.T, files map[string]string) drivers.RepoStore {
	root := t.TempDir()
	handle, err := drivers.Open("file", root, zap.NewNop())
	require.NoError(t, err)

	repo, ok := handle.RepoStore()
	require.True(t, ok)

	for path, data := range files {
		repo.Put(context.Background(), "", path, strings.NewReader(data))
	}

	return repo
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
