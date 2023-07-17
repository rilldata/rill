package rillv1

import (
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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
-- @connector: postgres
-- @refresh.cron: 0 0 * * *
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

	truth := true
	resources := []*Resource{
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
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s2"},
			Paths: []string{"/sources/s2.sql"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "postgres",
				Properties:      must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["sources/s2.sql"])})),
				RefreshSchedule: &runtimev1.Schedule{Cron: "0 0 * * *"},
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
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
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

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", []string{""})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestLocationError(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// source s1
		`sources/s1.yaml`: `
connector: s3
path: hello
  world: foo
`,
		// model m1
		`/models/m1.sql`: `
-- @materialize: true
SELECT *

FRO m1
`,
	}

	errors := []*runtimev1.ParseError{
		{
			Message:       " mapping values are not allowed in this context",
			FilePath:      "/sources/s1.yaml",
			StartLocation: &runtimev1.CharLocation{Line: 4},
		},
		{
			Message:       "syntax error at or near",
			FilePath:      "/models/m1.sql",
			StartLocation: &runtimev1.CharLocation{Line: 5},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", []string{""})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, errors)
}

func TestReparse(t *testing.T) {
	// Prepare
	truth := true
	ctx := context.Background()

	// Create empty project
	repo := makeRepo(t, map[string]string{`rill.yaml`: ``})
	p, err := Parse(ctx, repo, "", []string{""})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, nil)

	// Add a source
	putRepo(t, repo, map[string]string{
		`sources/s1.yaml`: `
connector: s3
path: hello
`,
	})
	s1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
		Paths: []string{"/sources/s1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
		},
	}
	diff, err := p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1}, nil)
	require.Equal(t, &Diff{
		Added: []ResourceName{s1.Name},
	}, diff)

	// Add a model
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
SELECT * FROM foo
`,
	})
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Refs:  []ResourceName{{Name: "foo"}},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Sql: "SELECT * FROM foo",
		},
	}
	diff, err = p.Reparse(ctx, m1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Added: []ResourceName{m1.Name},
	}, diff)

	// Annotate the model with a YAML file
	putRepo(t, repo, map[string]string{
		`models/m1.yaml`: `
materialize: true
`,
	})
	m1.Paths = []string{"/models/m1.sql", "/models/m1.yaml"}
	m1.ModelSpec.Materialize = &truth
	diff, err = p.Reparse(ctx, []string{"/models/m1.yaml"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name},
	}, diff)

	// Modify the model's SQL
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
SELECT * FROM bar
`,
	})
	m1.Refs = []ResourceName{{Name: "bar"}}
	m1.ModelSpec.Sql = "SELECT * FROM bar"
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name},
	}, diff)

	// Add a syntax error in the source
	putRepo(t, repo, map[string]string{
		`sources/s1.yaml`: `
connector: s3
path: hello
  world: path
`,
	})
	diff, err = p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, []*runtimev1.ParseError{{
		Message:       "mapping values are not allowed in this context", // note: approximate string match
		FilePath:      "/sources/s1.yaml",
		StartLocation: &runtimev1.CharLocation{Line: 4},
	}})
	require.Equal(t, &Diff{
		Deleted: []ResourceName{s1.Name},
	}, diff)

	// Delete the source
	deleteRepo(t, repo, s1.Paths[0])
	diff, err = p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)
	require.Equal(t, &Diff{}, diff)
}

func BenchmarkReparse(b *testing.B) {
	ctx := context.Background()
	truth := true
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
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
	}
	resources := []*Resource{
		// m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Sql: strings.TrimSpace(files["models/m1.sql"]),
			},
		},
		// m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Refs:  []ResourceName{{Name: "m1"}},
			Paths: []string{"/models/m2.sql", "/models/m2.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				Sql:         strings.TrimSpace(files["models/m2.sql"]),
				Materialize: &truth,
			},
		},
	}
	repo := makeRepo(b, files)
	p, err := Parse(ctx, repo, "", []string{""})
	require.NoError(b, err)
	requireResourcesAndErrors(b, p, resources, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files[`models/m2.sql`] = fmt.Sprintf(`SELECT * FROM m1 LIMIT %d`, i)
		_, err = p.Reparse(ctx, []string{`models/m2.sql`})
		require.NoError(b, err)
		require.Empty(b, p.Errors)
	}
}

func TestEmbeddedSources(t *testing.T) {
	// Expected
	files := map[string]string{
		`rill.yaml`: ``,
		`sources/m1.sql`: `
SELECT * FROM "s3://bucket/path"
`,
		`models/m2.sql`: `
SELECT * FROM "s3://bucket/path"
`,
		`models/m3.sql`: `
SELECT * FROM m2
`,
	}
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Refs:  []ResourceName{{Kind: ResourceKindSource, Name: "embed_b3d6beea4bcd7d8ef3970707a4ddbabb"}},
		Paths: []string{"/sources/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Sql: "SELECT * FROM embed_b3d6beea4bcd7d8ef3970707a4ddbabb",
		},
	}
	m2 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
		Refs:  []ResourceName{{Kind: ResourceKindSource, Name: "embed_b3d6beea4bcd7d8ef3970707a4ddbabb"}},
		Paths: []string{"/models/m2.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Sql: "SELECT * FROM embed_b3d6beea4bcd7d8ef3970707a4ddbabb",
		},
	}
	m3 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m3"},
		Refs:  []ResourceName{{Name: "m2"}},
		Paths: []string{"/models/m3.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Sql: "SELECT * FROM m2",
		},
	}
	embed := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "embed_b3d6beea4bcd7d8ef3970707a4ddbabb"},
		Paths: []string{"/sources/m1.sql", "/models/m2.sql"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "s3://bucket/path"})),
		},
	}

	// Parse
	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", []string{""})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1, m2, m3, embed}, nil)

	// Delete m1
	embed.Paths = []string{"/models/m2.sql"}
	deleteRepo(t, repo, m1.Paths...)
	diff, err := p.Reparse(ctx, m1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m2, m3, embed}, nil)
	require.ElementsMatch(t, []ResourceName{}, diff.Added)
	require.ElementsMatch(t, []ResourceName{embed.Name, m2.Name}, diff.Modified)
	require.ElementsMatch(t, []ResourceName{m1.Name}, diff.Deleted)
}

func requireResourcesAndErrors(t testing.TB, p *Parser, wantResources []*Resource, wantErrors []*runtimev1.ParseError) {
	// Check resources
	gotResources := maps.Clone(p.Resources)
	for _, want := range wantResources {
		found := false
		for _, got := range gotResources {
			if want.Name == got.Name {
				require.Equal(t, want.Name, got.Name)
				require.ElementsMatch(t, want.Refs, got.Refs, "for resource %q", want.Name)
				require.ElementsMatch(t, want.Paths, got.Paths, "for resource %q", want.Name)
				require.Equal(t, want.SourceSpec, got.SourceSpec, "for resource %q", want.Name)
				require.Equal(t, want.ModelSpec, got.ModelSpec, "for resource %q", want.Name)
				require.Equal(t, want.MetricsViewSpec, got.MetricsViewSpec, "for resource %q", want.Name)
				require.Equal(t, want.MigrationSpec, got.MigrationSpec, "for resource %q", want.Name)

				delete(gotResources, got.Name)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing resource %v", want.Name)
		}
	}
	if len(gotResources) > 0 {
		t.Errorf("unexpected resources: %v", gotResources)
	}

	// Check errors
	// NOTE: Assumes there's at most one parse error per file path
	// NOTE: Matches error messages using Contains (exact match not required)
	gotErrors := slices.Clone(p.Errors)
	for _, want := range wantErrors {
		found := false
		for i, got := range gotErrors {
			if want.FilePath == got.FilePath {
				require.Contains(t, got.Message, want.Message, "for path %q", got.FilePath)
				require.Equal(t, want.StartLocation, got.StartLocation, "for path %q", got.FilePath)
				gotErrors = slices.Delete(gotErrors, i, i+1)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing error for path %q", want.FilePath)
		}
	}
	if len(gotErrors) > 0 {
		t.Errorf("unexpected errors: %v", gotErrors)
	}
}

func makeRepo(t testing.TB, files map[string]string) drivers.RepoStore {
	root := t.TempDir()
	handle, err := drivers.Open("file", map[string]any{"dsn": root}, zap.NewNop())
	require.NoError(t, err)

	repo, ok := handle.AsRepoStore()
	require.True(t, ok)

	putRepo(t, repo, files)

	return repo
}

func putRepo(t testing.TB, repo drivers.RepoStore, files map[string]string) {
	for path, data := range files {
		err := repo.Put(context.Background(), "", path, strings.NewReader(data))
		require.NoError(t, err)
	}
}

func deleteRepo(t testing.TB, repo drivers.RepoStore, files ...string) {
	for _, path := range files {
		err := repo.Delete(context.Background(), "", path)
		require.NoError(t, err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
