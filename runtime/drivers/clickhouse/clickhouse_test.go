package clickhouse

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func TestManagedModePrecedence(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]any
		expectManaged bool
		expectCleared bool
		expectError   bool
	}{
		{
			name: "managed true clears conflicting properties",
			config: map[string]any{
				"managed":  true,
				"username": "conflicting_user",
				"password": "conflicting_pass",
				"host":     "conflicting_host",
				"port":     9000,
				"database": "conflicting_db",
				"ssl":      true,
				"dsn":      "clickhouse://user:pass@host:9000/db",
			},
			expectManaged: true,
			expectCleared: true,
			expectError:   false,
		},
		{
			name: "managed false allows all properties",
			config: map[string]any{
				"managed":  false,
				"username": "test_user",
				"password": "test_pass",
				"host":     "test_host",
				"port":     9000,
				"database": "test_db",
				"ssl":      true,
			},
			expectManaged: false,
			expectCleared: false,
			expectError:   false,
		},
		{
			name: "managed true with no conflicting properties",
			config: map[string]any{
				"managed":        true,
				"log_queries":    true,
				"max_open_conns": 10,
			},
			expectManaged: true,
			expectCleared: false, // No properties to clear
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config properties
			conf := &configProperties{
				CanScaleToZero: true,
				MaxOpenConns:   20,
				MaxIdleConns:   5,
			}

			// Parse config using mapstructure (simulating the driver's behavior)
			err := mapstructure.WeakDecode(tt.config, conf)
			require.NoError(t, err)

			// Validate the config (this is where our changes take effect)
			err = conf.validate()

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectManaged, conf.Managed)

			if tt.expectCleared {
				// Verify that conflicting properties were cleared
				require.Empty(t, conf.Username, "username should be cleared in managed mode")
				require.Empty(t, conf.Password, "password should be cleared in managed mode")
				require.Empty(t, conf.Host, "host should be cleared in managed mode")
				require.Zero(t, conf.Port, "port should be cleared in managed mode")
				require.Empty(t, conf.Database, "database should be cleared in managed mode")
				require.False(t, conf.SSL, "ssl should be cleared in managed mode")
				require.Empty(t, conf.DSN, "dsn should be cleared in managed mode")
			}
		})
	}
}
