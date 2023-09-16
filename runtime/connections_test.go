package runtime_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestAcquireHandle(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	ctx := context.Background()

	// set some env variable in instance
	inst, err := rt.Instance(ctx, id)
	require.NoError(t, err)
	inst.Variables = map[string]string{"aws_secret_access_key": "yyyy"}
	require.NoError(t, rt.EditInstance(ctx, inst))

	repo, _, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	putRepo(t, repo, map[string]string{
		`rill.yaml`: `
title: Hello world
description: This project says hello to the world

connectors:
- name: my-s3
  type: s3
  defaults:
    aws_access_key_id: us-east-1
    aws_secret_access_key: xxxx

env:
  foo: bar
  allow_host_access: false
`,
	})

	handle, _, err := rt.AcquireHandle(ctx, id, "my-s3")
	require.NoError(t, err)
	config := handle.Config()
	require.True(t, config["aws_access_key_id"].(string) == "us-east-1")
	require.True(t, config["allow_host_access"].(bool))
}

func putRepo(t testing.TB, repo drivers.RepoStore, files map[string]string) {
	for path, data := range files {
		err := repo.Put(context.Background(), path, strings.NewReader(data))
		require.NoError(t, err)
	}
}
