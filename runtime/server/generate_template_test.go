package server_test

import (
	"fmt"
	"sort"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	// Register all connector drivers
	_ "github.com/rilldata/rill/runtime/drivers/athena"
	_ "github.com/rilldata/rill/runtime/drivers/azure"
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/clickhouse"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/https"
	_ "github.com/rilldata/rill/runtime/drivers/mysql"
	_ "github.com/rilldata/rill/runtime/drivers/pinot"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/redshift"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
	_ "github.com/rilldata/rill/runtime/drivers/salesforce"
	_ "github.com/rilldata/rill/runtime/drivers/snowflake"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/drivers/starrocks"
)

// TestPropertySpecSecretFlags asserts that PropertySpec.Secret flags match expected values.
func TestPropertySpecSecretFlags(t *testing.T) {
	expected := map[string][]string{
		"s3":         {"aws_access_key_id", "aws_secret_access_key", "aws_role_arn", "aws_role_session_name", "aws_external_id"},
		"gcs":        {"google_application_credentials", "key_id", "secret"},
		"azure":      {"azure_storage_account", "azure_storage_key", "azure_storage_sas_token", "azure_storage_connection_string"},
		"clickhouse": {"dsn", "password"},
		"postgres":   {"dsn", "password"},
		"bigquery":   {"google_application_credentials"},
		"snowflake":  {"dsn", "password"},
		"redshift":   {"aws_access_key_id", "aws_secret_access_key"},
		"motherduck": {"token"},
		"athena":     {"aws_access_key_id", "aws_secret_access_key"},
		"mysql":      {"dsn", "password"},
		"druid":      {"dsn", "password"},
		"pinot":      {"dsn", "password"},
		"starrocks":  {"dsn", "password"},
		"salesforce": {"password", "key"},
	}

	for driverName, drv := range drivers.Connectors {
		expectedKeys, ok := expected[driverName]
		if !ok {
			continue
		}
		spec := drv.Spec()
		actualSecrets := secretKeys(spec.ConfigProperties)
		sort.Strings(expectedKeys)
		sort.Strings(actualSecrets)
		require.Equal(t, expectedKeys, actualSecrets, "driver %s: secret keys mismatch", driverName)
	}
}

func secretKeys(props []*drivers.PropertySpec) []string {
	var keys []string
	for _, p := range props {
		if p.Secret {
			keys = append(keys, p.Key)
		}
	}
	return keys
}

// TestGenerateTemplate tests the GenerateTemplate RPC handler end-to-end.
func TestGenerateTemplate(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	tt := []struct {
		name         string
		req          *runtimev1.GenerateTemplateRequest
		wantContains []string
		wantExcludes []string
		wantEnvKeys  []string
		wantDriver   string
		wantResType  string
		wantErr      codes.Code
	}{
		{
			name: "clickhouse connector with parameters",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{"host": "ch.example.com", "port": float64(9000), "password": "secret123"}),
			},
			wantContains: []string{
				"type: connector",
				"driver: clickhouse",
				`host: "ch.example.com"`,
				"port: 9000",
				`{{ .env.CLICKHOUSE_PASSWORD }}`,
				"# Connector YAML",
			},
			wantExcludes: []string{"secret123"},
			wantEnvKeys:  []string{"CLICKHOUSE_PASSWORD"},
			wantDriver:   "clickhouse",
			wantResType:  "connector",
		},
		{
			name: "clickhouse connector with dsn",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{"dsn": "clickhouse://user:pass@host:9000/db"}),
			},
			wantContains: []string{`{{ .env.CLICKHOUSE_DSN }}`},
			wantExcludes: []string{"clickhouse://user:pass"},
			wantEnvKeys:  []string{"CLICKHOUSE_DSN"},
			wantDriver:   "clickhouse",
			wantResType:  "connector",
		},
		{
			name: "s3 connector",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "s3",
				Properties:   mustStruct(map[string]any{"aws_access_key_id": "AKIATEST", "aws_secret_access_key": "secretkey"}),
			},
			wantContains: []string{
				"driver: s3",
				`{{ .env.AWS_ACCESS_KEY_ID }}`,
				`{{ .env.AWS_SECRET_ACCESS_KEY }}`,
			},
			wantExcludes: []string{"AKIATEST", "secretkey"},
			wantEnvKeys:  []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
			wantDriver:   "s3",
			wantResType:  "connector",
		},
		{
			name: "bigquery connector",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "bigquery",
				Properties:   mustStruct(map[string]any{"project_id": "my-project", "google_application_credentials": `{"type":"service_account"}`}),
			},
			wantContains: []string{
				"driver: bigquery",
				`project_id: "my-project"`,
				`{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}`,
			},
			wantExcludes: []string{"service_account"},
			wantEnvKeys:  []string{"GOOGLE_APPLICATION_CREDENTIALS"},
			wantDriver:   "bigquery",
			wantResType:  "connector",
		},
		{
			name: "postgres connector with params",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "postgres",
				Properties:   mustStruct(map[string]any{"host": "db.example.com", "port": "5432", "password": "my_pg_secret"}),
			},
			wantContains: []string{
				"driver: postgres",
				`host: "db.example.com"`,
				`{{ .env.POSTGRES_PASSWORD }}`,
			},
			wantExcludes: []string{"my_pg_secret"},
			wantDriver:   "postgres",
			wantResType:  "connector",
		},
		{
			name: "empty values filtered",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{"host": "ch.example.com", "port": "", "database": ""}),
			},
			wantContains: []string{"host:"},
			wantExcludes: []string{"port:", "database:"},
			wantDriver:   "clickhouse",
			wantResType:  "connector",
		},
		{
			name: "clickhouse managed false excluded",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{"host": "ch.example.com", "managed": false}),
			},
			wantContains: []string{"host:"},
			wantExcludes: []string{"managed"},
			wantDriver:   "clickhouse",
			wantResType:  "connector",
		},
		{
			name: "s3 model rewritten to duckdb parquet",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:    instanceID,
				ResourceType:  "model",
				Driver:        "s3",
				Properties:    mustStruct(map[string]any{"path": "s3://bucket/data.parquet", "name": "my_source"}),
				ConnectorName: "my_s3",
			},
			wantContains: []string{
				"type: model",
				`connector: "duckdb"`,
				"read_parquet",
				`create_secrets_from_connectors: "my_s3"`,
			},
			wantDriver:  "duckdb",
			wantResType: "model",
		},
		{
			name: "s3 model rewritten to duckdb csv",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "model",
				Driver:       "s3",
				Properties:   mustStruct(map[string]any{"path": "s3://bucket/data.csv", "name": "test"}),
			},
			wantContains: []string{
				"read_csv",
				"auto_detect=true",
				"ignore_errors=1",
				"header=true",
			},
			wantDriver:  "duckdb",
			wantResType: "model",
		},
		{
			name: "gcs model rewritten with json",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "model",
				Driver:       "gcs",
				Properties:   mustStruct(map[string]any{"path": "gs://bucket/data.json", "name": "test"}),
			},
			wantContains: []string{
				"read_json",
				"auto_detect=true",
				"format='auto'",
			},
			wantDriver:  "duckdb",
			wantResType: "model",
		},
		{
			name: "https model defaults to json",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:    instanceID,
				ResourceType:  "model",
				Driver:        "https",
				Properties:    mustStruct(map[string]any{"path": "https://api.example.com/data", "name": "test"}),
				ConnectorName: "my_http",
			},
			wantContains: []string{
				"read_json",
				`create_secrets_from_connectors: "my_http"`,
			},
			wantDriver:  "duckdb",
			wantResType: "model",
		},
		{
			name: "local_file csv model",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "model",
				Driver:       "local_file",
				Properties:   mustStruct(map[string]any{"path": "/data/file.csv", "name": "test"}),
			},
			wantContains: []string{
				"read_csv",
			},
			wantExcludes: []string{"create_secrets_from_connectors"},
			wantDriver:   "duckdb",
			wantResType:  "model",
		},
		{
			name: "sqlite model rewritten",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "model",
				Driver:       "sqlite",
				Properties:   mustStruct(map[string]any{"db": "/data/app.db", "table": "users", "name": "test"}),
			},
			wantContains: []string{
				"type: model",
				"sqlite_scan('/data/app.db', 'users')",
			},
			wantDriver:  "duckdb",
			wantResType: "model",
		},
		{
			name: "clickhouse not rewritten for model",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:    instanceID,
				ResourceType:  "model",
				Driver:        "clickhouse",
				Properties:    mustStruct(map[string]any{"sql": "SELECT * FROM events"}),
				ConnectorName: "ch_prod",
			},
			wantContains: []string{
				"type: model",
				"materialize: true",
				`connector: "ch_prod"`,
				"SELECT * FROM events",
				"dev:",
				"limit 10000",
			},
			wantDriver:  "clickhouse",
			wantResType: "model",
		},
		{
			name: "redshift model without dev section",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:    instanceID,
				ResourceType:  "model",
				Driver:        "redshift",
				Properties:    mustStruct(map[string]any{"sql": "SELECT * FROM events"}),
				ConnectorName: "rs_prod",
			},
			wantContains: []string{"type: model", `connector: "rs_prod"`, "materialize: true"},
			wantExcludes: []string{"dev:"},
			wantDriver:   "redshift",
			wantResType:  "model",
		},
		{
			name: "unknown driver rejected",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "nonexistent",
				Properties:   mustStruct(map[string]any{}),
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "unknown property rejected",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{"bogus_key": "value"}),
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "invalid resource type rejected",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "dashboard",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{}),
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "secret values never in error messages",
			req: &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       "clickhouse",
				Properties:   mustStruct(map[string]any{"bogus_key": "super_secret_value"}),
			},
			wantErr: codes.InvalidArgument,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := srv.GenerateTemplate(ctx, tc.req)
			if tc.wantErr != 0 {
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.wantErr, s.Code())
				// Ensure secret values never appear in error messages
				require.NotContains(t, err.Error(), "super_secret_value")
				return
			}
			require.NoError(t, err)

			for _, c := range tc.wantContains {
				require.Contains(t, resp.Blob, c, "blob should contain %q", c)
			}
			for _, e := range tc.wantExcludes {
				require.NotContains(t, resp.Blob, e, "blob should not contain %q", e)
			}
			if tc.wantDriver != "" {
				require.Equal(t, tc.wantDriver, resp.Driver)
			}
			if tc.wantResType != "" {
				require.Equal(t, tc.wantResType, resp.ResourceType)
			}
			for _, envKey := range tc.wantEnvKeys {
				require.Contains(t, resp.EnvVars, envKey, "env_vars should contain key %q", envKey)
			}
		})
	}
}

// TestGenerateTemplateEnvConflict tests env var conflict resolution with existing .env files.
func TestGenerateTemplateEnvConflict(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			".env":      "CLICKHOUSE_PASSWORD=old_value\n",
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
		InstanceId:   instanceID,
		ResourceType: "connector",
		Driver:       "clickhouse",
		Properties:   mustStruct(map[string]any{"host": "ch.example.com", "password": "new_secret"}),
	})
	require.NoError(t, err)

	// Should use CLICKHOUSE_PASSWORD_1 since CLICKHOUSE_PASSWORD already exists
	require.Contains(t, resp.EnvVars, "CLICKHOUSE_PASSWORD_1")
	require.Contains(t, resp.Blob, "{{ .env.CLICKHOUSE_PASSWORD_1 }}")
	require.NotContains(t, resp.Blob, "new_secret")
}

// TestGenerateTemplateEnvConflictDouble tests double-conflict resolution.
func TestGenerateTemplateEnvConflictDouble(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			".env":      "CLICKHOUSE_PASSWORD=old\nCLICKHOUSE_PASSWORD_1=also_old\n",
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
		InstanceId:   instanceID,
		ResourceType: "connector",
		Driver:       "clickhouse",
		Properties:   mustStruct(map[string]any{"host": "ch.example.com", "password": "new_secret"}),
	})
	require.NoError(t, err)

	require.Contains(t, resp.EnvVars, "CLICKHOUSE_PASSWORD_2")
	require.Contains(t, resp.Blob, "{{ .env.CLICKHOUSE_PASSWORD_2 }}")
}

// TestGenerateTemplateMotherduckConnector tests MotherDuck connector generation.
func TestGenerateTemplateMotherduckConnector(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
		InstanceId:   instanceID,
		ResourceType: "connector",
		Driver:       "motherduck",
		Properties:   mustStruct(map[string]any{"path": "md:my_db", "token": "my_token_123"}),
	})
	require.NoError(t, err)

	require.Contains(t, resp.Blob, "driver: motherduck")
	require.Contains(t, resp.Blob, `{{ .env.MOTHERDUCK_TOKEN }}`)
	require.NotContains(t, resp.Blob, "my_token_123")
	require.Contains(t, resp.EnvVars, "MOTHERDUCK_TOKEN")
	require.Equal(t, "my_token_123", resp.EnvVars["MOTHERDUCK_TOKEN"])
}

// TestGenerateTemplateSnowflakeConnector tests Snowflake connector generation.
func TestGenerateTemplateSnowflakeConnector(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
		InstanceId:   instanceID,
		ResourceType: "connector",
		Driver:       "snowflake",
		Properties:   mustStruct(map[string]any{"account": "my_account", "user": "admin", "password": "pw123", "database": "analytics"}),
	})
	require.NoError(t, err)

	require.Contains(t, resp.Blob, "driver: snowflake")
	require.Contains(t, resp.Blob, `account: "my_account"`)
	require.Contains(t, resp.Blob, `{{ .env.SNOWFLAKE_PASSWORD }}`)
	require.NotContains(t, resp.Blob, "pw123")
}

// TestGenerateTemplateAllConnectorDrivers ensures every registered connector can generate a connector YAML.
func TestGenerateTemplateAllConnectorDrivers(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	// Drivers that have ConfigProperties (connector-type drivers)
	driversWithConfig := []struct {
		name  string
		props map[string]any
	}{
		{"s3", map[string]any{"aws_access_key_id": "test", "aws_secret_access_key": "test"}},
		{"gcs", map[string]any{"google_application_credentials": "{}"}},
		{"azure", map[string]any{"azure_storage_connection_string": "test"}},
		{"clickhouse", map[string]any{"host": "localhost"}},
		{"postgres", map[string]any{"host": "localhost"}},
		{"bigquery", map[string]any{"google_application_credentials": "{}"}},
		{"snowflake", map[string]any{"account": "test", "user": "test", "password": "test"}},
		{"redshift", map[string]any{"aws_access_key_id": "test", "aws_secret_access_key": "test"}},
		{"motherduck", map[string]any{"path": "md:test", "token": "test"}},
		{"athena", map[string]any{"aws_access_key_id": "test", "aws_secret_access_key": "test"}},
		{"mysql", map[string]any{"host": "localhost"}},
		{"druid", map[string]any{"host": "localhost"}},
		{"pinot", map[string]any{"broker_host": "localhost"}},
		{"starrocks", map[string]any{"host": "localhost"}},
		{"salesforce", map[string]any{"username": "test"}},
	}

	for _, d := range driversWithConfig {
		t.Run(d.name, func(t *testing.T) {
			resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       d.name,
				Properties:   mustStruct(d.props),
			})
			require.NoError(t, err)
			require.Contains(t, resp.Blob, "type: connector")
			require.Contains(t, resp.Blob, "driver: "+d.name)
		})
	}
}

// TestGenerateTemplateDuckDBRewrite tests all DuckDB rewrite cases.
func TestGenerateTemplateDuckDBRewrite(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	tt := []struct {
		name         string
		driver       string
		props        map[string]any
		connName     string
		wantDriver   string
		wantContains []string
		wantExcludes []string
	}{
		{
			name:       "s3 parquet",
			driver:     "s3",
			props:      map[string]any{"path": "s3://bucket/data.parquet", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_parquet('s3://bucket/data.parquet')",
			},
		},
		{
			name:       "s3 csv",
			driver:     "s3",
			props:      map[string]any{"path": "s3://bucket/data.csv", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_csv('s3://bucket/data.csv', auto_detect=true, ignore_errors=1, header=true)",
			},
		},
		{
			name:       "s3 compressed parquet",
			driver:     "s3",
			props:      map[string]any{"path": "s3://bucket/data.v1.parquet.gz", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_parquet('s3://bucket/data.v1.parquet.gz')",
			},
		},
		{
			name:       "gcs ndjson",
			driver:     "gcs",
			props:      map[string]any{"path": "gs://bucket/data.ndjson", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_json('gs://bucket/data.ndjson', auto_detect=true, format='auto')",
			},
		},
		{
			name:       "azure tsv",
			driver:     "azure",
			props:      map[string]any{"path": "azure://container/data.tsv", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_csv('azure://container/data.tsv', auto_detect=true, ignore_errors=1, header=true)",
			},
		},
		{
			name:       "s3 unknown extension",
			driver:     "s3",
			props:      map[string]any{"path": "s3://bucket/data.avro", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"select * from 's3://bucket/data.avro'",
			},
		},
		{
			name:         "s3 with connector name sets secrets",
			driver:       "s3",
			props:        map[string]any{"path": "s3://bucket/data.parquet", "name": "test"},
			connName:     "my_s3",
			wantDriver:   "duckdb",
			wantContains: []string{`create_secrets_from_connectors: "my_s3"`},
		},
		{
			name:       "https defaults to json",
			driver:     "https",
			props:      map[string]any{"path": "https://api.example.com/data", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_json('https://api.example.com/data', auto_detect=true, format='auto')",
			},
		},
		{
			name:       "https with csv extension",
			driver:     "https",
			props:      map[string]any{"path": "https://example.com/data.csv", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"read_csv('https://example.com/data.csv', auto_detect=true, ignore_errors=1, header=true)",
			},
		},
		{
			name:         "https with connector name",
			driver:       "https",
			props:        map[string]any{"path": "https://api.example.com/data", "name": "test"},
			connName:     "my_http",
			wantDriver:   "duckdb",
			wantContains: []string{`create_secrets_from_connectors: "my_http"`},
		},
		{
			name:         "local_file csv",
			driver:       "local_file",
			props:        map[string]any{"path": "/data/file.csv", "name": "test"},
			wantDriver:   "duckdb",
			wantContains: []string{"read_csv('/data/file.csv'"},
			wantExcludes: []string{"create_secrets_from_connectors"},
		},
		{
			name:       "sqlite",
			driver:     "sqlite",
			props:      map[string]any{"db": "/data/app.db", "table": "users", "name": "test"},
			wantDriver: "duckdb",
			wantContains: []string{
				"sqlite_scan('/data/app.db', 'users')",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
				InstanceId:    instanceID,
				ResourceType:  "model",
				Driver:        tc.driver,
				Properties:    mustStruct(tc.props),
				ConnectorName: tc.connName,
			})
			require.NoError(t, err)
			require.Equal(t, tc.wantDriver, resp.Driver)
			require.Equal(t, "model", resp.ResourceType)
			for _, c := range tc.wantContains {
				require.Contains(t, resp.Blob, c, "blob should contain %q", c)
			}
			for _, e := range tc.wantExcludes {
				require.NotContains(t, resp.Blob, e, "blob should not contain %q", e)
			}
		})
	}
}

// TestGenerateTemplateEnvVarNaming tests env var name resolution.
func TestGenerateTemplateEnvVarNaming(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	tt := []struct {
		name       string
		driver     string
		props      map[string]any
		wantEnvKey string
	}{
		{
			name:       "s3 access key uses EnvVarName",
			driver:     "s3",
			props:      map[string]any{"aws_access_key_id": "test"},
			wantEnvKey: "AWS_ACCESS_KEY_ID",
		},
		{
			name:       "bigquery creds uses EnvVarName",
			driver:     "bigquery",
			props:      map[string]any{"google_application_credentials": "{}"},
			wantEnvKey: "GOOGLE_APPLICATION_CREDENTIALS",
		},
		{
			name:       "motherduck token uses EnvVarName",
			driver:     "motherduck",
			props:      map[string]any{"token": "test", "path": "md:test"},
			wantEnvKey: "MOTHERDUCK_TOKEN",
		},
		{
			name:       "clickhouse password uses EnvVarName",
			driver:     "clickhouse",
			props:      map[string]any{"password": "test"},
			wantEnvKey: "CLICKHOUSE_PASSWORD",
		},
		{
			name:       "starrocks dsn uses fallback format",
			driver:     "starrocks",
			props:      map[string]any{"dsn": "test"},
			wantEnvKey: "STARROCKS_DSN",
		},
		{
			name:       "starrocks password uses fallback format",
			driver:     "starrocks",
			props:      map[string]any{"password": "test"},
			wantEnvKey: "STARROCKS_PASSWORD",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: "connector",
				Driver:       tc.driver,
				Properties:   mustStruct(tc.props),
			})
			require.NoError(t, err)
			require.Contains(t, resp.EnvVars, tc.wantEnvKey, "env_vars should contain %q", tc.wantEnvKey)
			require.Contains(t, resp.Blob, fmt.Sprintf("{{ .env.%s }}", tc.wantEnvKey))
		})
	}
}

// TestGenerateTemplateValidateProperties tests property validation.
func TestGenerateTemplateValidateProperties(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})

	ctx := testCtx()
	srv, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)

	tt := []struct {
		name         string
		driver       string
		resourceType string
		props        map[string]any
		wantErr      bool
	}{
		{
			name:         "valid connector props",
			driver:       "clickhouse",
			resourceType: "connector",
			props:        map[string]any{"host": "localhost"},
			wantErr:      false,
		},
		{
			name:         "unknown prop rejected",
			driver:       "clickhouse",
			resourceType: "connector",
			props:        map[string]any{"bogus": "value"},
			wantErr:      true,
		},
		{
			name:         "source prop on connector rejected",
			driver:       "duckdb",
			resourceType: "connector",
			props:        map[string]any{"sql": "SELECT 1"},
			wantErr:      true,
		},
		{
			name:         "source prop on model accepted",
			driver:       "duckdb",
			resourceType: "model",
			props:        map[string]any{"sql": "SELECT 1"},
			wantErr:      false,
		},
		{
			name:         "empty properties valid",
			driver:       "clickhouse",
			resourceType: "connector",
			props:        map[string]any{},
			wantErr:      false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.GenerateTemplate(ctx, &runtimev1.GenerateTemplateRequest{
				InstanceId:   instanceID,
				ResourceType: tc.resourceType,
				Driver:       tc.driver,
				Properties:   mustStruct(tc.props),
			})
			if tc.wantErr {
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, s.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func mustStruct(m map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	if err != nil {
		panic(err)
	}
	return s
}

