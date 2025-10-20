package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func getTestServer(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstance(t)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func testCtx() context.Context {
	return auth.WithClaims(context.Background(), &runtime.SecurityClaims{SkipChecks: true})
}
