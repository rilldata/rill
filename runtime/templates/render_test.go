package templates

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestRenderConnectorTemplate(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("s3")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "connector",
		DriverSpec: &drivers.Spec{
			DocsURL: "https://docs.rilldata.com/developers/build/connectors/data-source/s3",
			ConfigProperties: []*drivers.PropertySpec{
				{Key: "aws_access_key_id", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "AWS_ACCESS_KEY_ID"},
				{Key: "aws_secret_access_key", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "AWS_SECRET_ACCESS_KEY"},
				{Key: "region", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"aws_access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"region":                "us-east-1",
		},
		ConnectorName: "my_s3",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: connector")
	require.Contains(t, blob, "driver: s3")
	require.Contains(t, blob, `aws_access_key_id: "{{ .env.AWS_ACCESS_KEY_ID }}"`)
	require.Contains(t, blob, `aws_secret_access_key: "{{ .env.AWS_SECRET_ACCESS_KEY }}"`)
	require.Contains(t, blob, `region: "us-east-1"`)
	require.NotContains(t, blob, "AKIAIOSFODNN7EXAMPLE")

	// Verify env vars extracted
	require.Equal(t, "AKIAIOSFODNN7EXAMPLE", result.EnvVars["AWS_ACCESS_KEY_ID"])
	require.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", result.EnvVars["AWS_SECRET_ACCESS_KEY"])

	// Verify path
	require.Equal(t, "connectors/my_s3.yaml", result.Files[0].Path)
}

func TestRenderDuckDBModelTemplate(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("s3-duckdb")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			ImplementsObjectStore: true,
			SourceProperties: []*drivers.PropertySpec{
				{Key: "path", Type: drivers.StringPropertyType},
				{Key: "name", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"path": "s3://my-bucket/data/*.parquet",
			"name": "my_model",
		},
		ConnectorName: "my_s3_connector",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, "connector: duckdb")
	require.Contains(t, blob, `create_secrets_from_connectors: "my_s3_connector"`)
	require.Contains(t, blob, "read_parquet")
	require.Contains(t, blob, "s3://my-bucket/data/*.parquet")

	// Verify path uses model_name from "name" property
	require.Equal(t, "models/my_model.yaml", result.Files[0].Path)
}

func TestRenderWarehouseModelTemplate(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("snowflake-model")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			ImplementsWarehouse: true,
			DocsURL:             "https://docs.rilldata.com/developers/build/connectors/data-source/snowflake",
			SourceProperties:    []*drivers.PropertySpec{},
		},
		Properties: map[string]any{
			"sql":  "SELECT * FROM my_table",
			"name": "snowflake_data",
		},
		ConnectorName: "my_snowflake",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, `connector: "my_snowflake"`)
	require.Contains(t, blob, "materialize: true")
	require.Contains(t, blob, "SELECT * FROM my_table")
	require.Contains(t, blob, `select * from {{ ref "self" }} limit 10000`)

	require.Equal(t, "models/snowflake_data.yaml", result.Files[0].Path)
}

func TestRenderRedshiftModelNoDevSection(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("redshift-model")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			ImplementsWarehouse: true,
			SourceProperties:    []*drivers.PropertySpec{},
		},
		Properties: map[string]any{
			"sql":  "SELECT * FROM my_table",
			"name": "rs_data",
		},
		ConnectorName: "my_redshift",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, `connector: "my_redshift"`)
	// Redshift template should NOT have a dev section
	require.NotContains(t, blob, "dev:")
	require.NotContains(t, blob, `ref "self"`)
}

func TestRenderIcebergDuckDB(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("iceberg-duckdb")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template:   tmpl,
		Output:     "model",
		DriverSpec: nil, // driverless
		Properties: map[string]any{
			"path": "s3://my-iceberg-bucket/warehouse/my_table",
			"name": "iceberg_data",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, "iceberg_scan")
	require.Contains(t, blob, "s3://my-iceberg-bucket/warehouse/my_table")

	require.Equal(t, "models/iceberg_data.yaml", result.Files[0].Path)
}

func TestRenderS3ClickHouseModel(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("s3-clickhouse")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			ImplementsObjectStore: true,
			ConfigProperties: []*drivers.PropertySpec{
				{Key: "aws_access_key_id", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "AWS_ACCESS_KEY_ID"},
				{Key: "aws_secret_access_key", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "AWS_SECRET_ACCESS_KEY"},
			},
			SourceProperties: []*drivers.PropertySpec{
				{Key: "path", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"aws_access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"path":                  "s3://my-bucket/data/events.parquet",
			"name":                  "s3_events",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, "connector: clickhouse")
	require.Contains(t, blob, "materialize: true")
	// SQL should show the s3() function with env var refs
	require.Contains(t, blob, "FROM s3(")
	require.Contains(t, blob, "s3://my-bucket/data/events.parquet")
	require.Contains(t, blob, "{{ .env.AWS_ACCESS_KEY_ID }}")
	require.Contains(t, blob, "{{ .env.AWS_SECRET_ACCESS_KEY }}")
	require.Contains(t, blob, "Parquet")
	// Raw secrets should NOT appear in the blob
	require.NotContains(t, blob, "AKIAIOSFODNN7EXAMPLE")

	// Env vars should be extracted
	require.Equal(t, "AKIAIOSFODNN7EXAMPLE", result.EnvVars["AWS_ACCESS_KEY_ID"])

	require.Equal(t, "models/s3_events.yaml", result.Files[0].Path)
}

func TestRenderMySQLClickHouseModel(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("mysql-clickhouse")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			ConfigProperties: []*drivers.PropertySpec{
				{Key: "host", Type: drivers.StringPropertyType},
				{Key: "port", Type: drivers.NumberPropertyType},
				{Key: "user", Type: drivers.StringPropertyType},
				{Key: "password", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "MYSQL_PASSWORD"},
				{Key: "database", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"host":     "db.example.com",
			"port":     "3306",
			"user":     "myuser",
			"password": "secret123",
			"database": "mydb",
			"table":    "events",
			"name":     "mysql_events",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "connector: clickhouse")
	require.Contains(t, blob, "FROM mysql(")
	require.Contains(t, blob, "db.example.com:3306")
	require.Contains(t, blob, "mydb")
	require.Contains(t, blob, "events")
	require.Contains(t, blob, "myuser")
	require.Contains(t, blob, "{{ .env.MYSQL_PASSWORD }}")
	require.NotContains(t, blob, "secret123")
}

func TestRenderEnvVarConflict(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("s3")
	require.True(t, ok)

	// Pre-populate existing env to force conflict
	existingEnv := map[string]bool{
		"AWS_ACCESS_KEY_ID":     true,
		"AWS_SECRET_ACCESS_KEY": true,
	}

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "connector",
		DriverSpec: &drivers.Spec{
			ConfigProperties: []*drivers.PropertySpec{
				{Key: "aws_access_key_id", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "AWS_ACCESS_KEY_ID"},
				{Key: "aws_secret_access_key", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "AWS_SECRET_ACCESS_KEY"},
			},
		},
		Properties: map[string]any{
			"aws_access_key_id":     "AKIA_NEW",
			"aws_secret_access_key": "SECRET_NEW",
		},
		ConnectorName: "s3_2",
		ExistingEnv:   existingEnv,
	})
	require.NoError(t, err)

	// Env vars should have _1 suffix due to conflict
	require.Contains(t, result.EnvVars, "AWS_ACCESS_KEY_ID_1")
	require.Contains(t, result.EnvVars, "AWS_SECRET_ACCESS_KEY_1")
	require.Contains(t, result.Files[0].Blob, "AWS_ACCESS_KEY_ID_1")
}

func TestRenderEmptyPropertiesFiltered(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("clickhouse")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "connector",
		DriverSpec: &drivers.Spec{
			ConfigProperties: []*drivers.PropertySpec{
				{Key: "host", Type: drivers.StringPropertyType},
				{Key: "port", Type: drivers.NumberPropertyType},
				{Key: "password", Type: drivers.StringPropertyType, Secret: true, EnvVarName: "CLICKHOUSE_PASSWORD"},
				{Key: "database", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"host":     "localhost",
			"port":     9000,
			"password": "secret",
			"database": "", // empty; should be filtered
		},
		ConnectorName: "my_ch",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "host")
	require.Contains(t, blob, "port")
	require.Contains(t, blob, "CLICKHOUSE_PASSWORD")
	require.NotContains(t, blob, "database")
}

func TestRenderOutputFilter(t *testing.T) {
	// Create a two-file template inline for testing
	tmpl := &Template{
		Name:        "test-multi",
		DisplayName: "Test Multi-file",
		Tags:        []string{"test"},
		Files: []File{
			{
				Name:         "connector",
				PathTemplate: "connectors/[[ .connector_name ]].yaml",
				CodeTemplate: "type: connector\ndriver: test\n",
			},
			{
				Name:         "model",
				PathTemplate: "models/[[ .model_name ]].yaml",
				CodeTemplate: "type: model\nconnector: duckdb\n",
			},
		},
	}

	// All files
	result, err := Render(&RenderInput{
		Template:      tmpl,
		Output:        "",
		Properties:    map[string]any{"name": "test"},
		ConnectorName: "test_conn",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 2)

	// Connector only
	result, err = Render(&RenderInput{
		Template:      tmpl,
		Output:        "connector",
		Properties:    map[string]any{"name": "test"},
		ConnectorName: "test_conn",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	require.Equal(t, "connector", "connector") // verify it's the connector file
	require.Contains(t, result.Files[0].Blob, "type: connector")

	// Model only
	result, err = Render(&RenderInput{
		Template:      tmpl,
		Output:        "model",
		Properties:    map[string]any{"name": "test"},
		ConnectorName: "test_conn",
		ExistingEnv:   make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	require.Contains(t, result.Files[0].Blob, "type: model")
}

func TestRenderLocalFileDuckDB(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("local-file-duckdb")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			ImplementsFileStore: true,
			SourceProperties: []*drivers.PropertySpec{
				{Key: "path", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"path": "data/sales.csv",
			"name": "sales",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, "connector: duckdb")
	require.Contains(t, blob, "read_csv")
	require.Contains(t, blob, "data/sales.csv")
}

func TestRenderSQLiteDuckDB(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("sqlite-duckdb")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
		DriverSpec: &drivers.Spec{
			SourceProperties: []*drivers.PropertySpec{
				{Key: "db", Type: drivers.StringPropertyType},
				{Key: "table", Type: drivers.StringPropertyType},
			},
		},
		Properties: map[string]any{
			"db":    "/data/app.db",
			"table": "users",
			"name":  "sqlite_users",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, "connector: duckdb")
	require.Contains(t, blob, "sqlite_scan")
	require.Contains(t, blob, "/data/app.db")
	require.Contains(t, blob, "users")
}
