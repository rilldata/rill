package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAssetsHandler(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":              "",
			"public/regions.geojson": `{"type":"FeatureCollection","features":[]}`,
			"private.geojson":        `{"type":"FeatureCollection","features":[]}`,
		},
	})

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	claims := &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ReadObjects}}
	call := func(path string, header http.Header) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(auth.WithClaims(context.Background(), claims))
		for k, v := range header {
			req.Header[k] = v
		}
		req.SetPathValue("instance_id", instanceID)
		req.SetPathValue("path", path)
		httputil.Handler(srv.assetsHandler).ServeHTTP(rec, req)
		return rec
	}

	// Serves files under a public path, with headers that keep them out of shared caches.
	rec := call("public/regions.geojson", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"type":"FeatureCollection","features":[]}`, rec.Body.String())
	require.Equal(t, "private, no-cache", rec.Header().Get("Cache-Control"))
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag)

	// Revalidates with the ETag.
	rec = call("public/regions.geojson", http.Header{"If-None-Match": {etag}})
	require.Equal(t, http.StatusNotModified, rec.Code)
	require.Empty(t, rec.Body.String())

	// Missing files are not found.
	rec = call("public/missing.geojson", nil)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// Files outside a public path are forbidden.
	rec = call("private.geojson", nil)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
