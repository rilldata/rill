package validate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsLocalProjectRunning(t *testing.T) {
	projectPath := t.TempDir()
	otherProjectPath := t.TempDir()
	configJSON := func(instanceID, path string) string {
		b, err := json.Marshal(map[string]string{"instance_id": instanceID, "project_path": path})
		require.NoError(t, err)
		return string(b)
	}
	validConfig := configJSON("default", projectPath)

	for _, tt := range []struct {
		name   string
		body   string
		status int
		want   bool
	}{
		{name: "same project", body: validConfig, want: true},
		{name: "different instance ID", body: configJSON("custom", projectPath), want: true},
		{name: "different project", body: configJSON("default", otherProjectPath)},
		{name: "unrelated service", body: "<html>VPN</html>"},
		{name: "empty config", body: "{}"},
		{name: "missing instance ID", body: configJSON("", projectPath)},
		{name: "missing project path", body: configJSON("default", "")},
		{name: "nonexistent server path", body: configJSON("default", filepath.Join(projectPath, "missing"))},
		{name: "non-OK response", body: validConfig, status: http.StatusNotFound},
		{name: "incomplete JSON", body: validConfig[:len(validConfig)-1]},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/local/config" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if tt.status != 0 {
					w.WriteHeader(tt.status)
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			require.Equal(t, tt.want, isLocalProjectRunning(t.Context(), projectPath, srv.URL))
		})
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validConfig))
	}))
	defer srv.Close()

	t.Run("relative project path", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		rel, err := filepath.Rel(cwd, projectPath)
		require.NoError(t, err)
		require.True(t, isLocalProjectRunning(t.Context(), rel, srv.URL))
	})
	t.Run("symlinked project path", func(t *testing.T) {
		alias := filepath.Join(t.TempDir(), "project")
		if err := os.Symlink(projectPath, alias); err != nil {
			t.Skipf("cannot create symlink: %v", err)
		}
		require.True(t, isLocalProjectRunning(t.Context(), alias, srv.URL))
	})
	t.Run("redirect", func(t *testing.T) {
		redirect := httptest.NewServer(http.RedirectHandler(srv.URL+"/local/config", http.StatusFound))
		defer redirect.Close()
		require.False(t, isLocalProjectRunning(t.Context(), projectPath, redirect.URL))
	})
	t.Run("unresponsive service", func(t *testing.T) {
		unresponsive := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer unresponsive.Close()
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()
		require.False(t, isLocalProjectRunning(ctx, projectPath, unresponsive.URL))
	})
	t.Run("closed port", func(t *testing.T) {
		closed := httptest.NewServer(http.NotFoundHandler())
		closed.Close()
		require.False(t, isLocalProjectRunning(t.Context(), projectPath, closed.URL))
	})
}
