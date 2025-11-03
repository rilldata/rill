package testruntime

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
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
		dsn := testclickhouse.Start(t)
		return map[string]string{"dsn": dsn, "mode": "readwrite"}
	},
	// clickhouse_cluster starts multiple test containers and configures them as a ClickHouse cluster.
	"clickhouse_cluster": func(t TestingT) map[string]string {
		dsn, cluster := testclickhouse.StartCluster(t)
		return map[string]string{"dsn": dsn, "cluster": cluster, "mode": "readwrite"}
	},
	// Bigquery connector connects to a real bigquery cluster using the credentials json in RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON.
	// The service account must have the following permissions:
	// - BigQuery Data Viewer
	// - BigQuery Job User
	// - BigQuery Read Session User
	// The test dataset is pre-populated with tables defined in testdata/init_data/bigquery_init_data.sql.
	"bigquery": func(t TestingT) map[string]string {
		loadDotEnv(t)
		gac := os.Getenv("RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON")
		require.NotEmpty(t, gac, "Bigquery RILL_RUNTIME_BIGQUERY_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON not configured")
		return map[string]string{"google_application_credentials": gac}
	},
	// Snowflake connector connects to a real snowflake cloud using dsn in RILL_RUNTIME_SNOWFLAKE_TEST_DSN
	// The test dataset is pre-populated with tables defined in testdata/init_data/snowflake_init_data.sql:
	"snowflake": func(t TestingT) map[string]string {
		loadDotEnv(t)
		dsn := os.Getenv("RILL_RUNTIME_SNOWFLAKE_TEST_DSN")
		require.NotEmpty(t, dsn, "RILL_RUNTIME_SNOWFLAKE_TEST_DSN not configured")
		return map[string]string{"dsn": dsn}
	},
	"motherduck": func(t TestingT) map[string]string {
		loadDotEnv(t)
		path := os.Getenv("RILL_RUNTIME_MOTHERDUCK_TEST_PATH")
		require.NotEmpty(t, path)
		token := os.Getenv("RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN")
		require.NotEmpty(t, token)

		return map[string]string{"path": path, "token": token}
	},
	// gcs connector uses an actual gcs bucket with data populated from testdata/init_data/azure.
	"gcs": func(t TestingT) map[string]string {
		loadDotEnv(t)
		gac := os.Getenv("RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON")
		require.NotEmpty(t, gac, "GCS RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON not configured")
		hmacKey := os.Getenv("RILL_RUNTIME_GCS_TEST_HMAC_KEY")
		hmacSecret := os.Getenv("RILL_RUNTIME_GCS_TEST_HMAC_SECRET")
		require.NotEmpty(t, hmacKey, "GCS RILL_RUNTIME_GCS_TEST_HMAC_KEY not configured")
		require.NotEmpty(t, hmacSecret, "GCS RILL_RUNTIME_GCS_TEST_HMAC_SECRET not configured")

		return map[string]string{
			"google_application_credentials": gac,
			"key_id":                         hmacKey,
			"secret":                         hmacSecret,
		}
	},
	"gcs_s3_compat": func(t TestingT) map[string]string {
		loadDotEnv(t)
		hmacKey := os.Getenv("RILL_RUNTIME_GCS_TEST_HMAC_KEY")
		hmacSecret := os.Getenv("RILL_RUNTIME_GCS_TEST_HMAC_SECRET")
		require.NotEmpty(t, hmacKey, "GCS RILL_RUNTIME_GCS_TEST_HMAC_KEY not configured")
		require.NotEmpty(t, hmacSecret, "GCS RILL_RUNTIME_GCS_TEST_HMAC_SECRET not configured")

		return map[string]string{
			"key_id": hmacKey,
			"secret": hmacSecret,
		}
	},
	// S3 connector uses an actual S3 bucket with data populated from testdata/init_data/azure.
	"s3": func(t TestingT) map[string]string {
		loadDotEnv(t)
		accessKeyID := os.Getenv("RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY")
		require.NotEmpty(t, accessKeyID, "S3 RILL_RUNTIME_S3_TEST_AWS_ACCESS_KEY_ID not configured")
		require.NotEmpty(t, secretAccessKey, "S3 RILL_RUNTIME_S3_TEST_AWS_SECRET_ACCESS_KEY not configured")
		return map[string]string{
			"aws_access_key_id":     accessKeyID,
			"aws_secret_access_key": secretAccessKey,
		}
	},
	// Athena connector connects to an actual Athena service.
	// The test dataset is pre-populated with table definitions in testdata/init_data/athena_init_data.sql,
	// and the actual data is stored on S3, which matches the data in testdata/init_data/azure/parquet_test.
	"athena": func(t TestingT) map[string]string {
		loadDotEnv(t)
		accessKeyID := os.Getenv("RILL_RUNTIME_ATHENA_TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("RILL_RUNTIME_ATHENA_TEST_AWS_SECRET_ACCESS_KEY")
		require.NotEmpty(t, accessKeyID, "Athena RILL_RUNTIME_ATHENA_TEST_AWS_ACCESS_KEY_ID not configured")
		require.NotEmpty(t, secretAccessKey, "Athena RILL_RUNTIME_ATHENA_TEST_AWS_SECRET_ACCESS_KEY not configured")
		return map[string]string{
			"aws_access_key_id":     accessKeyID,
			"aws_secret_access_key": secretAccessKey,
		}
	},
	// Redshift connector connects to an actual Redshift Serverless Service.
	// The test dataset is pre-populated with table definitions in testdata/init_data/redshift_init_data.sql,
	"redshift": func(t TestingT) map[string]string {
		loadDotEnv(t)
		accessKeyID := os.Getenv("RILL_RUNTIME_REDSHIFT_TEST_AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("RILL_RUNTIME_REDSHIFT_TEST_AWS_SECRET_ACCESS_KEY")
		require.NotEmpty(t, accessKeyID, "RILL_RUNTIME_REDSHIFT_TEST_AWS_ACCESS_KEY_ID not configured")
		require.NotEmpty(t, secretAccessKey, "RILL_RUNTIME_REDSHIFT_TEST_AWS_SECRET_ACCESS_KEY not configured")
		return map[string]string{
			"aws_access_key_id":     accessKeyID,
			"aws_secret_access_key": secretAccessKey,
		}
	},
	// druid connects to a real Druid cluster using the connection string in RILL_RUNTIME_DRUID_TEST_DSN.
	// This usually uses the master.in cluster.
	"druid": func(t TestingT) map[string]string {
		loadDotEnv(t)
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

		dsn := fmt.Sprintf("mysql://mysql:mysql@%v:%v/mysql", host, port.Port())
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
			"azure_storage_account":              azurite.AccountName,
		}
	},
	"pinot": func(t TestingT) map[string]string {
		ctx := context.Background()
		pinot, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "apachepinot/pinot:latest",
				ExposedPorts: []string{"9000/tcp", "8000/tcp"},
				Cmd:          []string{"QuickStart", "-type", "batch"},
				WaitingFor:   wait.ForLog("You can always go to http://localhost:9000").WithStartupTimeout(2 * time.Minute),
			},
			Started: true,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			err := pinot.Terminate(ctx)
			require.NoError(t, err)
		})

		host, err := pinot.Host(ctx)
		require.NoError(t, err)
		brokerPort, err := pinot.MappedPort(ctx, "8000")
		require.NoError(t, err)
		controllerPort, err := pinot.MappedPort(ctx, "9000")
		require.NoError(t, err)

		dsn := fmt.Sprintf("http://%s:%s?controller=http://%s:%s",
			host, brokerPort.Port(), host, controllerPort.Port())

		return map[string]string{"dsn": dsn}
	},
	"openai": func(t TestingT) map[string]string {
		loadDotEnv(t)
		apiKey := os.Getenv("RILL_RUNTIME_OPENAI_TEST_API_KEY")
		require.NotEmpty(t, apiKey)
		return map[string]string{"api_key": apiKey}
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

		return nil
	})
}

// loadDotEnv loads the .env file at the repo root (if any).
func loadDotEnv(t TestingT) {
	_, currentFile, _, _ := goruntime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		require.NoError(t, godotenv.Load(envPath))
	}
}
