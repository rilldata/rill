package resolvers

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestUnionSQLResolvers(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "union",
		ResolverProperties: map[string]any{
			"resolvers": []any{
				map[string]any{"name": "sql", "properties": map[string]any{"sql": "SELECT 1 AS a, 'x' AS b"}},
				map[string]any{"name": "sql", "properties": map[string]any{"sql": "SELECT 2 AS a, 'y' AS b"}},
			},
		},
		Claims: &runtime.SecurityClaims{SkipChecks: true},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]any
	for {
		row, err := res.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}

	require.Len(t, rows, 2)
	require.Equal(t, int32(1), rows[0]["a"])
	require.Equal(t, "x", rows[0]["b"])
	require.Equal(t, int32(2), rows[1]["a"])
	require.Equal(t, "y", rows[1]["b"])
}

func TestUnionDifferentSchemas(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "union",
		ResolverProperties: map[string]any{
			"resolvers": []any{
				map[string]any{"name": "sql", "properties": map[string]any{"sql": "SELECT 1 AS a"}},
				map[string]any{"name": "sql", "properties": map[string]any{"sql": "SELECT 'hello' AS b"}},
			},
		},
		Claims: &runtime.SecurityClaims{SkipChecks: true},
	})
	require.NoError(t, err)
	defer res.Close()

	// Schema should contain fields from both resolvers
	schema := res.Schema()
	require.Len(t, schema.Fields, 2)
	require.Equal(t, "a", schema.Fields[0].Name)
	require.Equal(t, "b", schema.Fields[1].Name)

	var rows []map[string]any
	for {
		row, err := res.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}

	require.Len(t, rows, 2)
}

func TestUnionEmpty(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	_, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "union",
		ResolverProperties: map[string]any{
			"resolvers": []any{},
		},
		Claims: &runtime.SecurityClaims{SkipChecks: true},
	})
	require.ErrorContains(t, err, "at least one resolver")
}

func TestUnionPassesArgs(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "union",
		ResolverProperties: map[string]any{
			"resolvers": []any{
				map[string]any{"name": "sql", "properties": map[string]any{"sql": "SELECT '{{ .args.name }}' AS greeting"}},
			},
		},
		Args:   map[string]any{"name": "world"},
		Claims: &runtime.SecurityClaims{SkipChecks: true},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]any
	for {
		row, err := res.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		rows = append(rows, row)
	}

	require.Len(t, rows, 1)
	require.Equal(t, "world", rows[0]["greeting"])
}
