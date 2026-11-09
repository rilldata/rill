package drivers_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestServiceCredentialMarker(t *testing.T) {
	ctx := context.Background()
	require.False(t, drivers.UsesServiceCredential(ctx))

	ctx = drivers.WithServiceCredential(ctx)
	require.True(t, drivers.UsesServiceCredential(ctx))

	ctx = drivers.WithoutServiceCredential(ctx)
	require.False(t, drivers.UsesServiceCredential(ctx))
}
