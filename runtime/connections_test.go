package runtime_test

import (
	"context"
	"testing"

	_ "github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestAcquireHandle(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			`rill.yaml`: `
title: Hello world
description: This project says hello to the world

connectors:
- name: my-s3
  type: s3
  defaults:
    aws_access_key_id: us-east-1
    aws_secret_access_key: xxxx

variables:
  foo: bar
  allow_host_access: false
`,
		},
		Variables: map[string]string{
			"aws_secret_access_key": "yyyy",
		},
	})
	ctx := context.Background()

	handle, _, err := rt.AcquireHandle(ctx, id, "my-s3")
	require.NoError(t, err)
	config := handle.Config()
	require.True(t, config["aws_access_key_id"].(string) == "us-east-1")
	require.True(t, config["allow_host_access"].(bool))
}
