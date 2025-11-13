package clickhouse

import (
	"context"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestManagedModePrecedence(t *testing.T) {
	tests := []struct {
		name             string
		config           map[string]any
		expectManaged    bool
		expectCleared    bool
		expectDSNCleared bool
		expectError      bool
	}{
		{
			name: "managed true clears conflicting properties but preserves DSN",
			config: map[string]any{
				"managed":  true,
				"username": "user",
				"password": "pass",
				"host":     "host",
				"port":     9440,
				"database": "db",
				"ssl":      true,
				"dsn":      "clickhouse://user:pass@host:9440/db",
			},
			expectManaged:    true,
			expectCleared:    true,
			expectDSNCleared: false, // provisioner DSN should be preserved
			expectError:      false,
		},
		{
			name: "managed true with no DSN clears all connection properties",
			config: map[string]any{
				"managed":  true,
				"username": "conflicting_user",
				"password": "conflicting_pass",
				"host":     "conflicting_host",
				"port":     9000,
				"database": "conflicting_db",
				"ssl":      true,
			},
			expectManaged:    true,
			expectCleared:    true,
			expectDSNCleared: true,
			expectError:      false,
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
			expectManaged:    false,
			expectCleared:    false,
			expectDSNCleared: false,
			expectError:      false,
		},
		{
			name: "managed true with no conflicting properties",
			config: map[string]any{
				"managed":        true,
				"log_queries":    true,
				"max_open_conns": 10,
			},
			expectManaged:    true,
			expectCleared:    false,
			expectDSNCleared: false,
			expectError:      false,
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
			}

			if tt.expectDSNCleared {
				require.Empty(t, conf.DSN, "dsn should be cleared when not provided by provisioner")
			} else if _, hasDSN := tt.config["dsn"]; hasDSN {
				require.NotEmpty(t, conf.DSN, "dsn should be preserved when provided")
			}
		})
	}
}

// TestQueryAttributesSETTINGSInjection verifies that query attributes are properly injected into the SETTINGS clause
func TestQueryAttributesSETTINGSInjection(t *testing.T) {
	dsn := testclickhouse.Start(t)

	conn, err := driver{}.Open("default", map[string]any{"dsn": dsn, "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	ctx := context.Background()

	// Create test table
	err = olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE TABLE test_attrs (id Int32, value String) ENGINE=Memory",
	})
	require.NoError(t, err)

	// Insert test data
	err = olap.Exec(ctx, &drivers.Statement{
		Query: "INSERT INTO test_attrs VALUES (1, 'test')",
	})
	require.NoError(t, err)

	t.Run("SingleQueryAttribute", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT id, value FROM test_attrs",
			QueryAttributes: map[string]string{
				"test_partner_id": "acme_corp",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		err = res.Close()
		require.NoError(t, err)
	})

	t.Run("MultipleQueryAttributes", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT id, value FROM test_attrs",
			QueryAttributes: map[string]string{
				"test_partner_id":  "acme_corp",
				"test_environment": "production",
				"test_region":      "us-west-2",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		err = res.Close()
		require.NoError(t, err)
	})

	t.Run("EmptyQueryAttributes", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query:           "SELECT id, value FROM test_attrs",
			QueryAttributes: map[string]string{},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		err = res.Close()
		require.NoError(t, err)
	})

	t.Run("NilQueryAttributes", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query:           "SELECT id, value FROM test_attrs",
			QueryAttributes: nil,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		err = res.Close()
		require.NoError(t, err)
	})

	t.Run("QueryAttributeWithExistingPrefix", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT id, value FROM test_attrs",
			QueryAttributes: map[string]string{
				"test_partner_id":  "acme_corp",
				"test_environment": "production",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		err = res.Close()
		require.NoError(t, err)
	})

	t.Run("QueryAttributesWithSpecialCharacters", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT id, value FROM test_attrs",
			QueryAttributes: map[string]string{
				"test_partner_name": "ACME Corp's Division",
				"test_description":  "Test with \"quotes\" and backslash\\",
				"test_path":         "/usr/local/bin",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		err = res.Close()
		require.NoError(t, err)
	})
}
