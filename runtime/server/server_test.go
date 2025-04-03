package server_test

import (
	"context"

	"github.com/rilldata/rill/runtime/server/auth"
)

func testCtx() context.Context {
	return auth.WithOpen(context.Background())
}
