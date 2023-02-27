package server

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func getTestServer(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstance(t)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}

func testCtx() context.Context {
	return auth.WithOpen(context.Background())
}
