package testruntime

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

// AcquireConnector acquires a test connector by name.
// For a list of available connectors, see the Connectors map below.
func AcquireConnector(t TestingT, name string) map[string]any {
	acquire, ok := Connectors[name]
	require.True(t, ok, "connector not found")
	vars := acquire(t)
	cfg := make(map[string]any, len(vars))
	for k, v := range vars {
		cfg[k] = v
	}
	return cfg
}

// ConnectorAcquireFunc is a function that acquires a connector for a test.
// It should return a map of config keys suitable for passing to drivers.Open.
type ConnectorAcquireFunc func(t TestingT) (vars map[string]string)

// Connectors is a map of available connectors for use in tests.
// When acquiring a connector, it will only be cleaned up when the test has completed.
// You should avoid acquiring the same connector multiple times in the same test.
//
// Test connectors can either be implemented as:
// - Services embedded in the current process
// - Services started as ephemeral testcontainers
// - Real external services configured for use in tests with credentials provided in the root .env file with the prefix RILL_RUNTIME_TEST_.
var Connectors = map[string]ConnectorAcquireFunc{
	// clickhouse starts a ClickHouse test container with no tables initialized.
	"clickhouse": func(t TestingT) map[string]string {
		_, currentFile, _, _ := goruntime.Caller(0)
		testdataPath := filepath.Join(currentFile, "..", "testdata")

		ctx := context.Background()
		clickHouseContainer, err := clickhouse.Run(
			ctx,
			"clickhouse/clickhouse-server:24.6.2.17",
			clickhouse.WithUsername("clickhouse"),
			clickhouse.WithPassword("clickhouse"),
			clickhouse.WithConfigFile(filepath.Join(testdataPath, "clickhouse-config.xml")),
			testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
				cf := testcontainers.ContainerFile{
					HostFilePath:      filepath.Join(testdataPath, "users.xml"),
					ContainerFilePath: "/etc/clickhouse-server/users.xml",
					FileMode:          0o755,
				}
				req.Files = append(req.Files, cf)
				return nil
			}),
		)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := clickHouseContainer.Terminate(ctx)
			require.NoError(t, err)
		})

		host, err := clickHouseContainer.Host(ctx)
		require.NoError(t, err)
		port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
		require.NoError(t, err)

		dsn := fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port())
		return map[string]string{"dsn": dsn}
	},
	// Bigquery connector connects to a real bigquery cluster using the credentials json in RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON.
	// The service account must have the following permissions:
	// - BigQuery Data Viewer
	// - BigQuery Job User
	// - BigQuery Read Session User
	// The test dataset is pre-populated with tables defined in testdata/init_data/bigquery_init_data.sql:

	"bigquery": func(t TestingT) map[string]string {
		// Load .env file at the repo root (if any)
		_, currentFile, _, _ := goruntime.Caller(0)
		envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
		_, err := os.Stat(envPath)
		if err == nil {
			require.NoError(t, godotenv.Load(envPath))
		}

		gac := os.Getenv("RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON")
		require.NotEmpty(t, gac, "Bigquery RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON not configured")
		return map[string]string{"google_application_credentials": gac}
	},
	"gcs": func(t TestingT) map[string]string {
		// Load .env file at the repo root (if any)
		_, currentFile, _, _ := goruntime.Caller(0)
		envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
		_, err := os.Stat(envPath)
		if err == nil {
			require.NoError(t, godotenv.Load(envPath))
		}
		hmacKey := os.Getenv("RILL_RUNTIME_GCS_TEST_HMAC_KEY")
		hmacSecret := os.Getenv("RILL_RUNTIME_GCS_TEST_HMAC_SECRET")
		gac := os.Getenv("RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON")
		require.NotEmpty(t, hmacKey, "GCS RILL_RUNTIME_GCS_TEST_HMAC_KEY not configured")
		require.NotEmpty(t, hmacSecret, "GCS RILL_RUNTIME_GCS_TEST_HMAC_SECRET not configured")
		require.NotEmpty(t, gac, "GCS RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON not configured")

		return map[string]string{
			"google_application_credentials": gac,
			"key_id":                         hmacKey,
			"secret":                         hmacSecret,
		}
	},
	"s3": func(t TestingT) map[string]string {
		// Load .env file at the repo root (if any)
		_, currentFile, _, _ := goruntime.Caller(0)
		envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
		_, err := os.Stat(envPath)
		if err == nil {
			require.NoError(t, godotenv.Load(envPath))
		}
		accessKeyID := os.Getenv("RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY")
		require.NotEmpty(t, accessKeyID, "S3 RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID not configured")
		require.NotEmpty(t, secretAccessKey, "S3 RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY not configured")
		return map[string]string{
			"aws_access_key_id":     accessKeyID,
			"aws_secret_access_key": secretAccessKey,
		}
	},
	// druid connects to a real Druid cluster using the connection string in RILL_RUNTIME_DRUID_TEST_DSN.
	// This usually uses the master.in cluster.
	"druid": func(t TestingT) map[string]string {
		// Load .env file at the repo root (if any)
		_, currentFile, _, _ := goruntime.Caller(0)
		envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
		_, err := os.Stat(envPath)
		if err == nil {
			require.NoError(t, godotenv.Load(envPath))
		}

		dsn := os.Getenv("RILL_RUNTIME_DRUID_TEST_DSN")
		require.NotEmpty(t, dsn, "Druid test DSN not configured")
		return map[string]string{"dsn": dsn}
	},
	"postgres": func(t TestingT) map[string]string {
		_, currentFile, _, _ := goruntime.Caller(0)
		testdataPath := filepath.Join(currentFile, "..", "testdata")
		postgresInitData := filepath.Join(testdataPath, "init_data", "postgres_init_data.sql")

		pgc := pgtestcontainer.New(t.(*testing.T))
		t.Cleanup(func() {
			pgc.Terminate(t.(*testing.T))
		})

		db, err := sql.Open("pgx", pgc.DatabaseURL)
		require.NoError(t, err)
		defer db.Close()
		sqlFile, err := os.ReadFile(postgresInitData)
		require.NoError(t, err)
		_, err = db.Exec(string(sqlFile))
		require.NoError(t, err)

		ip, err := pgc.Container.ContainerIP(context.Background())
		require.NoError(t, err)

		return map[string]string{
			"dsn": pgc.DatabaseURL,
			"ip":  ip,
		}
	},
	"mysql": func(t TestingT) map[string]string {
		_, currentFile, _, _ := goruntime.Caller(0)
		testdataPath := filepath.Join(currentFile, "..", "testdata")
		mysqlInitData := filepath.Join(testdataPath, "init_data", "mysql_init_data.sql")

		ctx := context.Background()
		mysqlContainer, err := mysql.Run(ctx,
			"mysql:8.0.36",
			mysql.WithUsername("mysql"),
			mysql.WithPassword("mysql"),
			mysql.WithDatabase("mysql"),
			mysql.WithScripts(mysqlInitData),
		)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := mysqlContainer.Terminate(ctx)
			require.NoError(t, err)
		})

		host, err := mysqlContainer.Host(ctx)
		require.NoError(t, err)
		port, err := mysqlContainer.MappedPort(ctx, "3306/tcp")
		require.NoError(t, err)

		dsn := fmt.Sprintf("mysql:mysql@tcp(%v:%v)/mysql", host, port.Port())
		ip, err := mysqlContainer.ContainerIP(context.Background())
		require.NoError(t, err)

		return map[string]string{"dsn": dsn, "ip": ip}
	},
	"azure": func(t TestingT) map[string]string {
		ctx := context.Background()
		azuriteContainer, err := azurite.Run(
			ctx,
			"mcr.microsoft.com/azure-storage/azurite:3.34.0",
			azurite.WithInMemoryPersistence(64),
		)
		t.Cleanup(func() {
			err := testcontainers.TerminateContainer(azuriteContainer)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		blobServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.BlobService), azurite.AccountName)

		cred, err := azblob.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
		require.NoError(t, err)
		client, err := azblob.NewClientWithSharedKeyCredential(blobServiceURL, cred, nil)
		require.NoError(t, err)
		containerName := "integration-test"
		_, err = client.CreateContainer(ctx, containerName, nil)
		require.NoError(t, err)

		_, currentFile, _, _ := goruntime.Caller(0)
		testdataPath := filepath.Join(currentFile, "..", "testdata")
		azureInitData := filepath.Join(testdataPath, "init_data", "azure")
		err = uploadDirectory(ctx, client, containerName, azureInitData)
		require.NoError(t, err)

		connectionString := fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=%s;AccountKey=%s;BlobEndpoint=%s;", azurite.AccountName, azurite.AccountKey, blobServiceURL)

		ip, err := azuriteContainer.ContainerIP(context.Background())
		require.NoError(t, err)
		blobEndpointWithIP := fmt.Sprintf("http://%s:%d/%s", ip, 10000, azurite.AccountName)

		connectionStringWithIP := fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=%s;AccountKey=%s;BlobEndpoint=%s;", azurite.AccountName, azurite.AccountKey, blobEndpointWithIP)

		return map[string]string{
			"azure_storage_connection_string":    connectionString,
			"azure_storage_connection_string_ip": connectionStringWithIP,
		}
	},
}

func uploadDirectory(ctx context.Context, client *azblob.Client, containerName, localDir string) error {
	return filepath.WalkDir(localDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get relative path for blob name (preserving directory structure)
		blobName, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		_, err = client.UploadFile(ctx, containerName, blobName, file, nil)
		if err != nil {
			return err
		}

		fmt.Printf("Uploaded: %s\n", blobName)
		return nil
	})
}
