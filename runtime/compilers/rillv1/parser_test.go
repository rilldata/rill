package rillv1

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
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

vars:
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
    column: a
measures:
  - name: b
    expression: count(*)
first_day_of_week: 7
first_month_of_year: 3
available_time_ranges:
  - P2W
  - range: P4W
  - range: P2M
    comparison_offsets:
      - P1M
      - offset: P4M
        range: P2M
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
				Connector: "duckdb",
				Version:   2,
				Sql:       strings.TrimSpace(files["init.sql"]),
			},
		},
		// source s1
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
			Paths: []string{"/sources/s1.yaml"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "s3",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// source s2
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s2"},
			Paths: []string{"/sources/s2.sql"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "postgres",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["sources/s2.sql"])})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true, Cron: "0 0 * * *"},
			},
		},
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["models/m1.sql"]),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// model m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/models/m2.yaml", "/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["models/m2.sql"]),
				Materialize:     &truth,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// dashboard d1
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m2"}},
			Paths: []string{"/dashboards/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector: "duckdb",
				Table:     "m2",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", Expression: "count(*)"},
				},
				FirstDayOfWeek:   7,
				FirstMonthOfYear: 3,
				AvailableTimeRanges: []*runtimev1.MetricsViewSpec_AvailableTimeRange{
					{Range: "P2W"},
					{Range: "P4W"},
					{
						Range: "P2M",
						ComparisonOffsets: []*runtimev1.MetricsViewSpec_AvailableComparisonOffset{
							{Offset: "P1M"},
							{Offset: "P4M", Range: "P2M"},
						},
					},
				},
			},
		},
		// migration c1
		{
			Name:  ResourceName{Kind: ResourceKindMigration, Name: "c1"},
			Paths: []string{"/custom/c1.yml"},
			MigrationSpec: &runtimev1.MigrationSpec{
				Connector: "duckdb",
				Version:   3,
				Sql:       "CREATE TABLE a(a integer);",
			},
		},
		// model c2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "c2"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m2"}},
			Paths: []string{"/custom/c2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["custom/c2.sql"]),
				Materialize:     &truth,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				UsesTemplating:  true,
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
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
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, nil, errors)
}

func TestUniqueSourceModelName(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// source s1
		`sources/s1.yaml`: `
connector: s3
`,
		// model s1
		`/models/s1.sql`: `
SELECT 1
`,
	}

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
			Paths: []string{"/sources/s1.yaml"},
			SourceSpec: &runtimev1.SourceSpec{
				SourceConnector: "s3",
				SinkConnector:   "duckdb",
				Properties:      must(structpb.NewStruct(map[string]any{})),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
	}

	errors := []*runtimev1.ParseError{
		{
			Message:  "model name collides with source \"s1\"",
			FilePath: "/models/s1.sql",
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, errors)
}

func TestReparse(t *testing.T) {
	// Prepare
	truth := true
	ctx := context.Background()

	// Create empty project
	repo := makeRepo(t, map[string]string{`rill.yaml`: ``})
	p, err := Parse(ctx, repo, "", "", "duckdb")
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
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
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
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT * FROM foo",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
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
	m1.ModelSpec.Sql = "SELECT * FROM bar"
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1, m1}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name},
	}, diff)

	// Rename the model to collide with the source
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
-- @name: s1
SELECT * FROM bar
`,
	})
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1}, []*runtimev1.ParseError{
		{
			Message:  "model name collides with source \"s1\"",
			FilePath: "/models/m1.sql",
		},
		{
			Message:  "model name collides with source \"s1\"",
			FilePath: "/models/m1.yaml",
		},
	})
	require.Equal(t, &Diff{
		Deleted: []ResourceName{m1.Name},
	}, diff)

	// Put m1 back and add a syntax error in the source
	putRepo(t, repo, map[string]string{
		`models/m1.sql`: `
SELECT * FROM bar
`,
		`sources/s1.yaml`: `
connector: s3
path: hello
  world: path
`,
	})
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql", "/sources/s1.yaml"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, []*runtimev1.ParseError{{
		Message:       "mapping values are not allowed in this context", // note: approximate string match
		FilePath:      "/sources/s1.yaml",
		StartLocation: &runtimev1.CharLocation{Line: 4},
	}})
	require.Equal(t, &Diff{
		Added:   []ResourceName{m1.Name},
		Deleted: []ResourceName{s1.Name},
	}, diff)

	// Delete the source
	deleteRepo(t, repo, s1.Paths[0])
	diff, err = p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)
	require.Equal(t, &Diff{}, diff)
}

func TestReparseSourceModelCollision(t *testing.T) {
	// Create project with model m1
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`models/m1.sql`: `
SELECT 10
		`,
	})
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT 10",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)

	// Add colliding source m1
	putRepo(t, repo, map[string]string{
		`sources/m1.yaml`: `
connector: s3
path: hello
`,
	})
	s1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "m1"},
		Paths: []string{"/sources/m1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	diff, err := p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1}, []*runtimev1.ParseError{
		{
			Message:  "model name collides with source \"m1\"",
			FilePath: "/models/m1.sql",
		},
	})
	require.Equal(t, &Diff{
		Added:   []ResourceName{s1.Name},
		Deleted: []ResourceName{m1.Name},
	}, diff)

	// Remove colliding source, verify model is restored
	deleteRepo(t, repo, "/sources/m1.yaml")
	diff, err = p.Reparse(ctx, s1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1}, nil)
	require.Equal(t, &Diff{
		Added:   []ResourceName{m1.Name},
		Deleted: []ResourceName{s1.Name},
	}, diff)
}

func TestReparseNameCollision(t *testing.T) {
	// Create project with model m1
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`models/m1.sql`: `
SELECT 10
		`,
		`models/nested/m1.sql`: `
SELECT 20
		`,
		`models/m2.sql`: `
SELECT * FROM m1
		`,
	})
	m1 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT 10",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	m1Nested := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/nested/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT 20",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	m2 := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
		Paths: []string{"/models/m2.sql"},
		Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT * FROM m1",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1, m2}, []*runtimev1.ParseError{
		{
			Message:  "name collision",
			FilePath: "/models/nested/m1.sql",
			External: true,
		},
	})

	// Remove colliding model, verify things still work
	deleteRepo(t, repo, "/models/m1.sql")
	diff, err := p.Reparse(ctx, m1.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{m1Nested, m2}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{m1.Name, m2.Name}, // m2 due to ref re-inference
	}, diff)
}

func TestReparseMultiKindNameCollision(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`:            ``,
		`models/m1.sql`:        `SELECT 10`,
		`models/nested/m1.sql`: `SELECT 20`,
		`sources/m1.yaml`: `
type: s3
path: hello
`,
	})
	src := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "m1"},
		Paths: []string{"/sources/m1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	mdl := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT 10",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{src}, []*runtimev1.ParseError{
		{
			Message:  "collides with source",
			FilePath: "/models/m1.sql",
			External: true,
		},
		{
			Message:  "name collision",
			FilePath: "/models/nested/m1.sql",
			External: true,
		},
	})

	// Delete source m1
	deleteRepo(t, repo, "/sources/m1.yaml")
	diff, err := p.Reparse(ctx, src.Paths)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, []*runtimev1.ParseError{
		{
			Message:  "name collision",
			FilePath: "/models/nested/m1.sql",
			External: true,
		},
	})
	require.Equal(t, &Diff{
		Added:   []ResourceName{mdl.Name},
		Deleted: []ResourceName{src.Name},
	}, diff)
}

func TestReparseRillYAML(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{})

	mdl := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
		Paths: []string{"/models/m1.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT 10",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	perr := &runtimev1.ParseError{
		Message:  "rill.yaml not found",
		FilePath: "/rill.yaml",
	}

	// Parse empty project. Expect rill.yaml error.
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	require.Nil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, nil, []*runtimev1.ParseError{perr})

	// Add rill.yaml. Expect success.
	putRepo(t, repo, map[string]string{
		`rill.yaml`: ``,
	})
	diff, err := p.Reparse(ctx, []string{"/rill.yaml"})
	require.NoError(t, err)
	require.True(t, diff.Reloaded)
	require.NotNil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, nil, nil)

	// Remove rill.yaml and add a model. Expect reloaded.
	deleteRepo(t, repo, "/rill.yaml")
	putRepo(t, repo, map[string]string{"/models/m1.sql": "SELECT 10"})
	diff, err = p.Reparse(ctx, []string{"/rill.yaml", "/models/m1.sql"})
	require.NoError(t, err)
	require.True(t, diff.Reloaded)
	require.Nil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, []*runtimev1.ParseError{perr})

	// Edit model. Expect nothing to happen because rill.yaml is still broken.
	putRepo(t, repo, map[string]string{"/models/m1.sql": "SELECT 20"})
	diff, err = p.Reparse(ctx, []string{"/models/m1.sql"})
	require.NoError(t, err)
	require.Equal(t, &Diff{Skipped: true}, diff)
	require.Nil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, []*runtimev1.ParseError{perr})

	// Fix rill.yaml. Expect reloaded.
	mdl.ModelSpec.Sql = "SELECT 20"
	putRepo(t, repo, map[string]string{"/rill.yaml": ""})
	diff, err = p.Reparse(ctx, []string{"/rill.yaml"})
	require.NoError(t, err)
	require.True(t, diff.Reloaded)
	require.NotNil(t, p.RillYAML)
	requireResourcesAndErrors(t, p, []*Resource{mdl}, nil)
}

func TestRefInferrence(t *testing.T) {
	// Create model referencing "bar"
	foo := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "foo"},
		Paths: []string{"/models/foo.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT * FROM bar",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model foo
		`models/foo.sql`: `SELECT * FROM bar`,
	})
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{foo}, nil)

	// Add model "bar"
	foo.Refs = []ResourceName{{Kind: ResourceKindModel, Name: "bar"}}
	bar := &Resource{
		Name:  ResourceName{Kind: ResourceKindModel, Name: "bar"},
		Paths: []string{"/models/bar.sql"},
		ModelSpec: &runtimev1.ModelSpec{
			Connector:       "duckdb",
			Sql:             "SELECT * FROM baz",
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}
	putRepo(t, repo, map[string]string{
		`models/bar.sql`: `SELECT * FROM baz`,
	})
	diff, err := p.Reparse(ctx, []string{"/models/bar.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{foo, bar}, nil)
	require.Equal(t, &Diff{
		Added:    []ResourceName{bar.Name},
		Modified: []ResourceName{foo.Name},
	}, diff)

	// Remove "bar"
	foo.Refs = nil
	deleteRepo(t, repo, bar.Paths[0])
	diff, err = p.Reparse(ctx, []string{"/models/bar.sql"})
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{foo}, nil)
	require.Equal(t, &Diff{
		Modified: []ResourceName{foo.Name},
		Deleted:  []ResourceName{bar.Name},
	}, diff)
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
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["models/m1.sql"]),
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/models/m2.sql", "/models/m2.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["models/m2.sql"]),
				Materialize:     &truth,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
	}
	repo := makeRepo(b, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
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

func TestProjectModelDefaults(t *testing.T) {
	ctx := context.Background()
	truth := true
	falsity := false

	files := map[string]string{
		// Provide dashboard defaults in rill.yaml
		`rill.yaml`: `
models:
  materialize: true
`,
		// Model that inherits defaults
		`models/m1.sql`: `
SELECT * FROM t1
`,
		// Model that overrides defaults
		`models/m2.sql`: `
-- @materialize: false
SELECT * FROM t2
`,
	}

	resources := []*Resource{
		// m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["models/m1.sql"]),
				Materialize:     &truth,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		// m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Paths: []string{"/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             strings.TrimSpace(files["models/m2.sql"]),
				Materialize:     &falsity,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
	}

	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestProjectDashboardDefaults(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		// Provide dashboard defaults in rill.yaml
		`rill.yaml`: `
dashboards:
  first_day_of_week: 7
  available_time_zones:
    - America/New_York
  security:
    access: true
`,
		// Dashboard that inherits defaults
		`dashboards/d1.yaml`: `
table: t1
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
`,
		// Dashboard that overrides defaults
		`dashboards/d2.yaml`: `
table: t2
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
first_day_of_week: 1
available_time_zones: []
security:
  row_filter: true
`,
	})

	resources := []*Resource{
		// dashboard d1
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Paths: []string{"/dashboards/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector: "duckdb",
				Table:     "t1",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", Expression: "count(*)"},
				},
				FirstDayOfWeek:     7,
				AvailableTimeZones: []string{"America/New_York"},
				Security: &runtimev1.MetricsViewSpec_SecurityV2{
					Access: "true",
				},
			},
		},
		// dashboard d2
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d2"},
			Paths: []string{"/dashboards/d2.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector: "duckdb",
				Table:     "t2",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", Expression: "count(*)"},
				},
				FirstDayOfWeek:     1,
				AvailableTimeZones: []string{},
				Security: &runtimev1.MetricsViewSpec_SecurityV2{
					Access:    "true",
					RowFilter: "true",
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestEnvironmentOverrides(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		// Provide dashboard defaults in rill.yaml
		`rill.yaml`: `
env:
  test:
    sources:
      limit: 10000
`,
		// source s1
		`sources/s1.yaml`: `
connector: s3
path: hello
env:
  test:
    path: world
    refresh:
      cron: "0 0 * * *"
`,
	})

	s1Base := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
		Paths: []string{"/sources/s1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "hello"})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
		},
	}

	s1Test := &Resource{
		Name:  ResourceName{Kind: ResourceKindSource, Name: "s1"},
		Paths: []string{"/sources/s1.yaml"},
		SourceSpec: &runtimev1.SourceSpec{
			SourceConnector: "s3",
			SinkConnector:   "duckdb",
			Properties:      must(structpb.NewStruct(map[string]any{"path": "world", "limit": 10000})),
			RefreshSchedule: &runtimev1.Schedule{RefUpdate: true, Cron: "0 0 * * *"},
		},
	}

	// Parse without environment
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1Base}, nil)

	// Parse in environment "test"
	p, err = Parse(ctx, repo, "", "test", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, []*Resource{s1Test}, nil)
}

func TestReport(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`reports/r1.yaml`: `
kind: report
title: My Report

refresh:
  cron: 0 * * * *
  time_zone: America/Los_Angeles

query:
  name: MetricsViewToplist
  args:
    metrics_view: mv1

export:
  format: csv
  limit: 10000

email:
  recipients:
    - benjamin@example.com

annotations:
  foo: bar
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindReport, Name: "r1"},
			Paths: []string{"/reports/r1.yaml"},
			ReportSpec: &runtimev1.ReportSpec{
				Title: "My Report",
				RefreshSchedule: &runtimev1.Schedule{
					RefUpdate: true,
					Cron:      "0 * * * *",
					TimeZone:  "America/Los_Angeles",
				},
				QueryName:       "MetricsViewToplist",
				QueryArgsJson:   `{"metrics_view":"mv1"}`,
				ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
				ExportLimit:     10000,
				EmailRecipients: []string{"jane@example.com"},
				Annotations:     map[string]string{"foo": "bar"},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAlert(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1`,
		`alerts/a1.yaml`: `
kind: alert
title: My Alert

refs:
  - model/m1

refresh:
  ref_update: false
  cron: '0 * * * *'

watermark: inherit

intervals:
  duration: PT1H
  limit: 10

query:
  name: MetricsViewToplist
  args:
    metrics_view: mv1
  for:
    user_email: benjamin@example.com

email:
  on_recover: true
  renotify: true
  renotify_after: 24h
  recipients:
    - benjamin@example.com

annotations:
  foo: bar
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             `SELECT 1`,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAlert, Name: "a1"},
			Paths: []string{"/alerts/a1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			AlertSpec: &runtimev1.AlertSpec{
				Title: "My Alert",
				RefreshSchedule: &runtimev1.Schedule{
					RefUpdate:     false,
					TickerSeconds: 86400,
				},
				WatermarkInherit:          true,
				IntervalsIsoDuration:      "PT1H",
				IntervalsLimit:            10,
				QueryName:                 "MetricsViewToplist",
				QueryArgsJson:             `{"metrics_view":"mv1"}`,
				QueryFor:                  &runtimev1.AlertSpec_QueryForUserEmail{QueryForUserEmail: "benjamin@example.com"},
				EmailRecipients:           []string{"jane@example.com"},
				EmailOnRecover:            true,
				EmailOnFail:               true,
				EmailOnError:              false,
				EmailRenotify:             true,
				EmailRenotifyAfterSeconds: 24 * 60 * 60,
				Annotations:               map[string]string{"foo": "bar"},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestMetricsViewAvoidSelfCyclicRef(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// dashboard d1
		`dashboards/d1.yaml`: `
model: d1
dimensions:
  - name: a
    column: a
measures:
  - name: b
    expression: count(*)
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindMetricsView, Name: "d1"},
			Refs:  nil, // NOTE: This is what we're testing – that it avoids inferring the missing "d1" as a self-reference
			Paths: []string{"/dashboards/d1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector: "duckdb",
				Table:     "d1",
				Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
					{Name: "a", Column: "a"},
				},
				Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
					{Name: "b", Expression: "count(*)"},
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestTheme(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`themes/t1.yaml`: `
kind: theme

colors:
  primary: red
  secondary: grey
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindTheme, Name: "t1"},
			Paths: []string{"/themes/t1.yaml"},
			ThemeSpec: &runtimev1.ThemeSpec{
				PrimaryColor: &runtimev1.Color{
					Red:   1,
					Green: 0,
					Blue:  0,
					Alpha: 1,
				},
				SecondaryColor: &runtimev1.Color{
					Red:   0.5019608,
					Green: 0.5019608,
					Blue:  0.5019608,
					Alpha: 1,
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestChartsAndDashboard(t *testing.T) {
	vegaLiteSpec := `
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "description": "A simple bar chart with embedded data.",
    "mark": "bar",
    "data": {
      "name": "table"
    },
    "encoding": {
      "x": {"field": "time", "type": "nominal", "axis": {"labelAngle": 0}},
      "y": {"field": "total_sales", "type": "quantitative"}
    }
  }`
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`charts/c1.yaml`: fmt.Sprintf(`
kind: chart
data:
  api: MetricsViewAggregation
  args:
    metrics_view: foo
vega_lite: |%s
`, vegaLiteSpec),
		`charts/c2.yaml`: fmt.Sprintf(`
kind: chart
data:
  api: MetricsViewAggregation
  args:
    metrics_view: bar
vega_lite: |%s
`, vegaLiteSpec),
		`dashboards/d1.yaml`: `
kind: dashboard
grid:
  rows: 3
  columns: 4
components:
  - chart: c1
  - chart: c2
    rows: 1
    columns: 2
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindChart, Name: "c1"},
			Paths: []string{"/charts/c1.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindAPI, Name: "MetricsViewAggregation"}},
			ChartSpec: &runtimev1.ChartSpec{
				Resolver:           "api",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"api": "MetricsViewAggregation", "args": map[string]any{"metrics_view": "foo"}})),
				VegaLiteSpec:       vegaLiteSpec,
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindChart, Name: "c2"},
			Paths: []string{"/charts/c2.yaml"},
			Refs:  []ResourceName{{Kind: ResourceKindAPI, Name: "MetricsViewAggregation"}},
			ChartSpec: &runtimev1.ChartSpec{
				Resolver:           "api",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"api": "MetricsViewAggregation", "args": map[string]any{"metrics_view": "bar"}})),
				VegaLiteSpec:       vegaLiteSpec,
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindDashboard, Name: "d1"},
			Paths: []string{"/dashboards/d1.yaml"},
			Refs: []ResourceName{
				{Kind: ResourceKindChart, Name: "c1"},
				{Kind: ResourceKindChart, Name: "c2"},
			},
			DashboardSpec: &runtimev1.DashboardSpec{
				Grid: &runtimev1.Grid{Gap: 3, Columns: 4},
				Components: []*runtimev1.DashboardComponent{
					{Chart: "c1"},
					{Chart: "c2", Width: 1, Height: 2},
				},
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAPI(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1`,
		// api a1
		`apis/a1.yaml`: `
kind: api
sql: select * from m1
`,
		// api a2
		`apis/a2.yaml`: `
kind: api
metrics_sql: select * from m1
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				Connector:       "duckdb",
				Sql:             `SELECT 1`,
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a1"},
			Paths: []string{"/apis/a1.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a2"},
			Paths: []string{"/apis/a2.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
			},
		},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
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
				require.Equal(t, want.ThemeSpec, got.ThemeSpec, "for resource %q", want.Name)

				delete(gotResources, got.Name)
				found = true
				break
			}
		}
		require.True(t, found, "missing resource %q", want.Name)
	}
	require.True(t, len(gotResources) == 0, "unexpected resources: %v", gotResources)

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
		require.True(t, found, "missing error for path %q", want.FilePath)
	}
	require.True(t, len(gotErrors) == 0, "unexpected errors: %v", gotErrors)
}

func makeRepo(t testing.TB, files map[string]string) drivers.RepoStore {
	root := t.TempDir()
	handle, err := drivers.Open("file", map[string]any{"dsn": root}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	repo, ok := handle.AsRepoStore("")
	require.True(t, ok)

	putRepo(t, repo, files)

	return repo
}

func putRepo(t testing.TB, repo drivers.RepoStore, files map[string]string) {
	for path, data := range files {
		err := repo.Put(context.Background(), path, strings.NewReader(data))
		require.NoError(t, err)
	}
}

func deleteRepo(t testing.TB, repo drivers.RepoStore, files ...string) {
	for _, path := range files {
		err := repo.Delete(context.Background(), path)
		require.NoError(t, err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
