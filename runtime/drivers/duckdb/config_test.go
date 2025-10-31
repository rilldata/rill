package duckdb

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig(map[string]any{})
	require.NoError(t, err)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "", "cpu": 2})
	require.NoError(t, err)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"pool_size": 10})
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db", "memory_limit_gb": "8", "cpu": "2"})
	require.NoError(t, err)
	require.Equal(t, 2, cfg.PoolSize)
}

func Test_specialCharInPath(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "let's t@st \"weird\" dirs")
	err := os.Mkdir(path, fs.ModePerm)
	require.NoError(t, err)

	dbFile := filepath.Join(path, "st@g3's.db")
	conn, err := Driver{}.Open("default", map[string]any{"init_sql": fmt.Sprintf("ATTACH %s", safeSQLString(dbFile))}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT 1"})
	require.NoError(t, err)
	require.NoError(t, res.Close())
	require.NoError(t, conn.Close())
}

func TestModeConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
	}{
		{
			name:        "valid read mode",
			config:      map[string]any{"mode": modeReadOnly},
			expectError: false,
		},
		{
			name:        "valid readwrite mode",
			config:      map[string]any{"mode": modeReadWrite},
			expectError: false,
		},
		{
			name:        "empty mode is valid",
			config:      map[string]any{},
			expectError: false,
		},
		{
			name:        "invalid mode should fail",
			config:      map[string]any{"mode": "invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newConfig(tt.config)
			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid mode")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigManaged(t *testing.T) {
	tests := []struct {
		name           string
		configMap      map[string]any
		expectedMode   string
		expectedPath   string
		expectedAttach string
		expectedErr    string
	}{
		{
			name: "managed false should preserve path and set read mode",
			configMap: map[string]any{
				"managed": false,
				"path":    "/tmp/test.db",
			},
			expectedMode:   modeReadOnly,
			expectedPath:   "/tmp/test.db",
			expectedAttach: "",
		},
		{
			name: "managed true without path/attach should default to readwrite mode",
			configMap: map[string]any{
				"managed": true,
			},
			expectedMode:   modeReadWrite,
			expectedPath:   "",
			expectedAttach: "",
		},
		{
			name:           "no managed flag, no path/attach should default to readwrite mode",
			configMap:      map[string]any{},
			expectedMode:   modeReadWrite,
			expectedPath:   "",
			expectedAttach: "",
		},
		{
			name: "no managed flag with path should default to read mode",
			configMap: map[string]any{
				"path": "/tmp/test.db",
			},
			expectedMode:   modeReadOnly,
			expectedPath:   "/tmp/test.db",
			expectedAttach: "",
		},
		{
			name: "no managed flag with attach should default to read mode",
			configMap: map[string]any{
				"attach": "'test.db' AS test",
			},
			expectedMode:   modeReadOnly,
			expectedPath:   "",
			expectedAttach: "'test.db' AS test",
		},
		{
			name: "explicit mode should override default behavior",
			configMap: map[string]any{
				"path": "/tmp/test.db",
				"mode": modeReadWrite,
			},
			expectedMode:   modeReadWrite,
			expectedPath:   "/tmp/test.db",
			expectedAttach: "",
		},
		{
			name: "managed true with path should error",
			configMap: map[string]any{
				"managed": true,
				"path":    "/tmp/test.db",
			},
			expectedErr: "'managed: true' cannot be combined with 'path' or 'attach' fields",
		},
		{
			name: "managed true with attach should error",
			configMap: map[string]any{
				"managed": true,
				"attach":  "'test.db' AS test",
			},
			expectedErr: "'managed: true' cannot be combined with 'path' or 'attach' fields",
		},
		{
			name: "managed true with both path and attach should error",
			configMap: map[string]any{
				"managed": true,
				"path":    "/tmp/test.db",
				"attach":  "'test.db' AS test",
			},
			expectedErr: "'managed: true' cannot be combined with 'path' or 'attach' fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newConfig(tt.configMap)
			if tt.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.expectedMode, cfg.Mode)
			require.Equal(t, tt.expectedPath, cfg.Path)
			require.Equal(t, tt.expectedAttach, cfg.Attach)
		})
	}
}
