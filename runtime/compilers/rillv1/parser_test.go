package rillv1

import (
	"context"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/drivers/file"
)

func TestRillYAML(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t,
		`rill.yaml`,
		`
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
	)

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

func makeRepo(t *testing.T, pathsAndContents ...string) drivers.RepoStore {
	require.Equal(t, 0, len(pathsAndContents)%2)

	root := t.TempDir()
	handle, err := drivers.Open("file", root, zap.NewNop())
	require.NoError(t, err)

	repo, ok := handle.RepoStore()
	require.True(t, ok)

	for i := 0; i < len(pathsAndContents); i += 2 {
		repo.Put(context.Background(), "", pathsAndContents[i], strings.NewReader(strings.TrimSpace(pathsAndContents[i+1])))
	}

	return repo
}
