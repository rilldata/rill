package server

import (
	"context"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func getTestServer(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstance(t)

	server, err := NewServer(context.Background(), &Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func testCtx() context.Context {
	return auth.WithOpen(context.Background())
}
