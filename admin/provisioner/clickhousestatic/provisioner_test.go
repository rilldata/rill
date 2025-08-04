package clickhousestatic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainersclickhouse "github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"go.uber.org/zap"
)

func TestClickHouseStatic(t *testing.T) {
	// Create a test ClickHouse cluster
	container, err := testcontainersclickhouse.Run(
		context.Background(),
		"clickhouse/clickhouse-server:24.11.1.2557",
		// Add a user config file that enables access management for the "default" user
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.Files = append(req.Files, testcontainers.ContainerFile{
				Reader:            strings.NewReader(`<clickhouse><users><default><access_management>1</access_management></default></users></clickhouse>`),
				ContainerFilePath: "/etc/clickhouse-server/users.d/default.xml",
				FileMode:          0o755,
			})
			return nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	})
	host, err := container.Host(context.Background())
	require.NoError(t, err)
	port, err := container.MappedPort(context.Background(), "9000/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("clickhouse://default:default@%v:%v", host, port.Port())

	// Create the provisioner
	specJSON, err := json.Marshal(&Spec{
		DSN: dsn,
	})
	require.NoError(t, err)
	p, err := New(specJSON, nil, zap.NewNop())
	require.NoError(t, err)

	// Provision two resources
	r1, db1 := provisionClickHouse(t, p)
	defer db1.Close()
	r2, db2 := provisionClickHouse(t, p)
	defer db2.Close()

	// Check the resources are different
	require.NotEqual(t, r1.ID, r2.ID)
	require.NotEqual(t, r1.Config["dsn"], r2.Config["dsn"])

	// Create a table with the first connection
	_, err = db1.Exec("CREATE TABLE test (id UInt64) ENGINE = Memory")
	require.NoError(t, err)
	_, err = db1.Exec("INSERT INTO test VALUES (1)")
	require.NoError(t, err)
	rows, err := db1.Query("SELECT COUNT(*) FROM system.tables WHERE database <> 'system'")
	require.NoError(t, err)
	for rows.Next() {
		var count int
		err = rows.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, count, 1)
	}
	require.NoError(t, err)
	rows.Close()

	// Get the name of the first connection's database
	dsn1, err := clickhouse.ParseDSN(r1.Config["dsn"].(string))
	require.NoError(t, err)
	db1Name := dsn1.Auth.Database

	// Check the second connection doesn't have access to the table in the first connection
	_, err = db2.Exec(fmt.Sprintf("SELECT * FROM %s.test", db1Name))
	require.Error(t, err)
	_, err = db2.Exec("SELECT * FROM test")
	require.Error(t, err)

	// Check the second connection can't see the other connection's tables in the information schema
	rows, err = db2.Query("SELECT name FROM system.tables WHERE database <> 'system'")
	require.NoError(t, err)
	for rows.Next() {
		require.Fail(t, "unexpected visible table in information schema")
	}
	require.NoError(t, err)
	rows.Close()

	// Deprovision the resources
	err = p.Deprovision(context.Background(), r1)
	require.NoError(t, err)
	err = p.Deprovision(context.Background(), r2)
	require.NoError(t, err)

	// Check the connections are deficient
	_, err = db1.Exec("SELECT 1")
	require.Error(t, err)
	_, err = db2.Exec("SELECT 1")
	require.Error(t, err)
}

func provisionClickHouse(t *testing.T, p provisioner.Provisioner) (*provisioner.Resource, *sql.DB) {
	// Provision a new resource
	in := &provisioner.Resource{
		ID:     uuid.New().String(),
		Type:   provisioner.ResourceTypeClickHouse,
		State:  nil,
		Config: nil,
	}
	opts := &provisioner.ResourceOptions{
		Args:        nil,
		Annotations: map[string]string{"organization": "test", "project": "test"},
		RillVersion: "dev",
	}
	out, err := p.Provision(context.Background(), in, opts)
	require.NoError(t, err)

	// Check the resource
	require.Equal(t, in.ID, out.ID)
	require.Equal(t, in.Type, out.Type)
	require.Empty(t, out.State)
	require.NotEmpty(t, out.Config)

	// Check the resource
	_, err = p.CheckResource(context.Background(), out, opts)
	require.NoError(t, err)

	// Open a connection to the database
	db, err := sql.Open("clickhouse", out.Config["dsn"].(string))
	require.NoError(t, err)

	// Ping
	err = db.Ping()
	require.NoError(t, err)

	return out, db
}

func TestClickHouseStaticWithEnvVar(t *testing.T) {
	// Create a test ClickHouse cluster
	container, err := testcontainersclickhouse.Run(
		context.Background(),
		"clickhouse/clickhouse-server:24.11.1.2557",
		// Add a user config file that enables access management for the "default" user
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.Files = append(req.Files, testcontainers.ContainerFile{
				Reader:            strings.NewReader(`<clickhouse><users><default><access_management>1</access_management></default></users></clickhouse>`),
				ContainerFilePath: "/etc/clickhouse-server/users.d/default.xml",
				FileMode:          0o755,
			})
			return nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	})
	host, err := container.Host(context.Background())
	require.NoError(t, err)
	port, err := container.MappedPort(context.Background(), "9000/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("clickhouse://default:default@%v:%v", host, port.Port())

	// Set environment variable
	envVar := "TEST_CLICKHOUSE_DSN"
	err = os.Setenv(envVar, dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Unsetenv(envVar)
	})

	// Create the provisioner using environment variable
	specJSON, err := json.Marshal(&Spec{
		DSNEnv: envVar,
	})
	require.NoError(t, err)
	p, err := New(specJSON, nil, zap.NewNop())
	require.NoError(t, err)

	// Provision a resource
	r, db := provisionClickHouse(t, p)
	defer db.Close()

	// Verify the resource works
	_, err = db.Exec("CREATE TABLE test (id UInt64) ENGINE = Memory")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO test VALUES (1)")
	require.NoError(t, err)

	// Cleanup
	err = p.Deprovision(context.Background(), r)
	require.NoError(t, err)
}

func TestClickHouseStaticEnvVarNotSet(t *testing.T) {
	// Test with environment variable that doesn't exist
	specJSON, err := json.Marshal(&Spec{
		DSNEnv: "NONEXISTENT_CLICKHOUSE_DSN",
	})
	require.NoError(t, err)
	_, err = New(specJSON, nil, zap.NewNop())
	require.Error(t, err)
	require.Contains(t, err.Error(), "environment variable \"NONEXISTENT_CLICKHOUSE_DSN\" is not set or empty")
}

func TestClickHouseStaticEnvVarEmpty(t *testing.T) {
	// Test with empty environment variable
	envVar := "EMPTY_CLICKHOUSE_DSN"
	err := os.Setenv(envVar, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Unsetenv(envVar)
	})

	specJSON, err := json.Marshal(&Spec{
		DSNEnv: envVar,
	})
	require.NoError(t, err)
	_, err = New(specJSON, nil, zap.NewNop())
	require.Error(t, err)
	require.Contains(t, err.Error(), "environment variable \"EMPTY_CLICKHOUSE_DSN\" is not set or empty")
}

func TestClickHouseStaticHumanReadableNaming(t *testing.T) {
	// Create a test ClickHouse cluster
	container, err := testcontainersclickhouse.Run(
		context.Background(),
		"clickhouse/clickhouse-server:24.11.1.2557",
		// Add a user config file that enables access management for the "default" user
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.Files = append(req.Files, testcontainers.ContainerFile{
				Reader:            strings.NewReader(`<clickhouse><users><default><access_management>1</access_management></default></users></clickhouse>`),
				ContainerFilePath: "/etc/clickhouse-server/users.d/default.xml",
				FileMode:          0o755,
			})
			return nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	})
	host, err := container.Host(context.Background())
	require.NoError(t, err)
	port, err := container.MappedPort(context.Background(), "9000/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("clickhouse://default:default@%v:%v", host, port.Port())

	// Create the provisioner
	specJSON, err := json.Marshal(&Spec{
		DSN: dsn,
	})
	require.NoError(t, err)
	p, err := New(specJSON, nil, zap.NewNop())
	require.NoError(t, err)

	// Test with org and project annotations
	resourceID := uuid.New().String()
	in := &provisioner.Resource{
		ID:     resourceID,
		Type:   provisioner.ResourceTypeClickHouse,
		State:  nil,
		Config: nil,
	}
	opts := &provisioner.ResourceOptions{
		Args: nil,
		Annotations: map[string]string{
			"organization_name": "Acme-Corp",
			"project_name":      "My-Project",
		},
		RillVersion: "dev",
	}

	out, err := p.Provision(context.Background(), in, opts)
	require.NoError(t, err)

	// Parse the DSN to get the database name and user
	cfg, err := provisioner.NewClickhouseConfig(out.Config)
	require.NoError(t, err)
	opts2, err := clickhouse.ParseDSN(cfg.DSN)
	require.NoError(t, err)
	// Check that the database name follows the expected format
	expectedID := strings.ReplaceAll(resourceID, "-", "")
	expectedDBName := fmt.Sprintf("rill_acme_corp_my_project_%s", expectedID)
	expectedUser := fmt.Sprintf("rill_%s", expectedID) // User name remains UUID format

	require.Equal(t, expectedDBName, opts2.Auth.Database)
	require.Equal(t, expectedUser, opts2.Auth.Username)

	// Test the connection works
	db, err := sql.Open("clickhouse", cfg.DSN)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	// Create a table to ensure permissions work
	_, err = db.Exec("CREATE TABLE test (id UInt64) ENGINE = Memory")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO test VALUES (1)")
	require.NoError(t, err)

	// Clean up
	err = p.Deprovision(context.Background(), out)
	require.NoError(t, err)
}

func TestClickHouseStaticFallbackNaming(t *testing.T) {
	// Create a test ClickHouse cluster
	container, err := testcontainersclickhouse.Run(
		context.Background(),
		"clickhouse/clickhouse-server:24.11.1.2557",
		// Add a user config file that enables access management for the "default" user
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.Files = append(req.Files, testcontainers.ContainerFile{
				Reader:            strings.NewReader(`<clickhouse><users><default><access_management>1</access_management></default></users></clickhouse>`),
				ContainerFilePath: "/etc/clickhouse-server/users.d/default.xml",
				FileMode:          0o755,
			})
			return nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	})
	host, err := container.Host(context.Background())
	require.NoError(t, err)
	port, err := container.MappedPort(context.Background(), "9000/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("clickhouse://default:default@%v:%v", host, port.Port())

	// Create the provisioner
	specJSON, err := json.Marshal(&Spec{
		DSN: dsn,
	})
	require.NoError(t, err)
	p, err := New(specJSON, nil, zap.NewNop())
	require.NoError(t, err)

	// Test without org/project annotations (should fall back to old format)
	resourceID := uuid.New().String()
	in := &provisioner.Resource{
		ID:     resourceID,
		Type:   provisioner.ResourceTypeClickHouse,
		State:  nil,
		Config: nil,
	}
	opts := &provisioner.ResourceOptions{
		Args:        nil,
		Annotations: map[string]string{}, // Empty annotations
		RillVersion: "dev",
	}

	out, err := p.Provision(context.Background(), in, opts)
	require.NoError(t, err)

	// Parse the DSN to get the database name and user
	cfg, err := provisioner.NewClickhouseConfig(out.Config)
	require.NoError(t, err)
	opts2, err := clickhouse.ParseDSN(cfg.DSN)
	require.NoError(t, err)
	// Check that the database name follows the fallback format
	expectedID := strings.ReplaceAll(resourceID, "-", "")
	expectedDBName := fmt.Sprintf("rill_%s", expectedID)
	expectedUser := fmt.Sprintf("rill_%s", expectedID) // User name always uses UUID format

	require.Equal(t, expectedDBName, opts2.Auth.Database)
	require.Equal(t, expectedUser, opts2.Auth.Username)

	// Log the database name for debugging
	t.Logf("Provisioned database name: %s", opts2.Auth.Database)
	t.Logf("Provisioned user name: %s", opts2.Auth.Username)

	// Test the connection works
	db, err := sql.Open("clickhouse", cfg.DSN)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	// Clean up
	err = p.Deprovision(context.Background(), out)
	require.NoError(t, err)
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"simple", "simple"},
		{"Simple", "simple"},
		{"UPPERCASE", "uppercase"},
		{"with-dashes", "with_dashes"},
		{"with spaces", "with_spaces"},
		{"with@special!chars", "withspecialchars"},
		{"mixed-Case_Name", "mixed_case_name"},
		{"123numbers", "123numbers"},
		{"_underscore_", "_underscore_"},
		{"Acme-Corp", "acme_corp"},
		{"My-Project", "my_project"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeName(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
