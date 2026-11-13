package executor

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

type oauthOLAPStub struct {
	drivers.OLAPStore
	cfg *drivers.OAuthConfig
}

func (s *oauthOLAPStub) OAuthConfig() *drivers.OAuthConfig {
	return s.cfg
}

func TestCacheKeyDisabledForPerUserOAuth(t *testing.T) {
	enabled := true
	spec := &runtimev1.MetricsViewSpec{CacheEnabled: &enabled}

	e := &Executor{metricsView: spec, olap: &oauthOLAPStub{cfg: &drivers.OAuthConfig{}}}
	_, ok, err := e.CacheKey(context.Background())
	require.NoError(t, err)
	require.False(t, ok)
	_, ok = e.TimestampsTTLCacheKey()
	require.False(t, ok)

	e = &Executor{metricsView: spec, olap: &oauthOLAPStub{}}
	_, ok, err = e.CacheKey(context.Background())
	require.NoError(t, err)
	require.True(t, ok)
	_, ok = e.TimestampsTTLCacheKey()
	require.True(t, ok)
}
