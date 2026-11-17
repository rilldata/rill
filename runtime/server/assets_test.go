package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAssetsHandler(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"></svg>`
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			// The root is listed to check that it never makes the whole project public.
			"rill.yaml":                  `public_paths: ["./public", "/static/", "/"]`,
			".env":                       "token=secret",
			"secret.txt":                 "secret",
			"publicity/notes.txt":        "secret",
			"public/logo.png":            "png",
			"public/img/logo.svg":        svg,
			"public/img/brand/PHOTO.JPG": "jpg",
			"public/fonts/brand.woff2":   "woff2",
			"public/page.html":           "<html></html>",
			"static/app.css":             "body {}",
		},
	})

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)
	handler, err := srv.HTTPHandler(context.Background(), nil, false)
	require.NoError(t, err)
	// Use a real server, so the paths go through the ServeMux's cleaning and unescaping like real requests do.
	httpSrv := httptest.NewServer(handler)
	t.Cleanup(httpSrv.Close)

	get := func(t *testing.T, path string) (*http.Response, string) {
		res, err := httpSrv.Client().Get(httpSrv.URL + "/v1/instances/" + instanceID + "/assets/" + path)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return res, string(body)
	}

	// Files in the public paths are served, at any depth, with a content type derived from the extension.
	served := []struct {
		path        string
		contentType string
		body        string
	}{
		{"public/logo.png", "image/png", "png"},
		{"public/img/logo.svg", "image/svg+xml", svg},
		{"public/img/brand/PHOTO.JPG", "image/jpeg", "jpg"},
		{"public/fonts/brand.woff2", "font/woff2", "woff2"},
		{"static/app.css", "text/css; charset=utf-8", "body {}"},
		// Types that aren't known to be passive are never served as something a browser renders.
		{"public/page.html", "application/octet-stream", "<html></html>"},
		// `..` segments that stay inside a public path are resolved.
		{"public/img/%2e%2e/logo.png", "image/png", "png"},
	}
	for _, tt := range served {
		t.Run(tt.path, func(t *testing.T) {
			res, body := get(t, tt.path)
			require.Equal(t, http.StatusOK, res.StatusCode)
			require.Equal(t, tt.contentType, res.Header.Get("Content-Type"))
			require.Equal(t, "nosniff", res.Header.Get("X-Content-Type-Options"))
			require.Contains(t, res.Header.Get("Content-Security-Policy"), "sandbox")
			require.Equal(t, tt.body, body)
		})
	}

	// Files outside the public paths are not served, however the path is spelled.
	forbidden := []string{
		"secret.txt",
		".env",
		"publicity/notes.txt",
		"public/../secret.txt",
		"public/%2e%2e/secret.txt",
		"public/%2E%2E/.env",
		"public/img/%2e%2e/%2e%2e/secret.txt",
		"public/..%2fsecret.txt",
		"public%2f..%2fsecret.txt",
		"public/./%2e%2e/secret.txt",
		"%2e%2e/secret.txt",
		"public/%2e%2e/publicity/notes.txt",
	}
	for _, path := range forbidden {
		t.Run(path, func(t *testing.T) {
			res, body := get(t, path)
			require.Equal(t, http.StatusForbidden, res.StatusCode)
			require.NotContains(t, body, "secret")
		})
	}

	// Missing files and directories are not served.
	res, _ := get(t, "public/img/missing.png")
	require.Equal(t, http.StatusNotFound, res.StatusCode)
	res, _ = get(t, "public/img")
	require.NotEqual(t, http.StatusOK, res.StatusCode)
}
