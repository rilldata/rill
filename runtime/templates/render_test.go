package templates

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestRenderConnectorTemplate(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	// s3-duckdb is a combined template with both connector and model files
	tmpl, ok := registry.Get("s3-duckdb")
	require.True(t, ok)

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "connector",
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

	// Connector output should NOT contain source-step properties
	require.NotContains(t, blob, "path")
	require.NotContains(t, blob, "name")

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

	// s3-duckdb has json_schema; no DriverSpec needed
	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "model",
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

	tmpl, ok := registry.Get("snowflake-duckdb")
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
	require.Contains(t, blob, "SELECT * FROM my_table limit 10000")

	require.Equal(t, "models/snowflake_data.yaml", result.Files[0].Path)
}

func TestRenderRedshiftModelNoDevSection(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("redshift-duckdb")
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

	// s3-duckdb is a combined template with json_schema
	tmpl, ok := registry.Get("s3-duckdb")
	require.True(t, ok)

	// Pre-populate existing env to force conflict
	existingEnv := map[string]bool{
		"AWS_ACCESS_KEY_ID":     true,
		"AWS_SECRET_ACCESS_KEY": true,
	}

	result, err := Render(&RenderInput{
		Template: tmpl,
		Output:   "connector",
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

	tmpl, ok := registry.Get("local_file-duckdb")
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

// --- Schema-based template tests ---

func TestRenderIcebergDuckDBWithSchema(t *testing.T) {
	registry, err := NewRegistry()
	require.NoError(t, err)

	tmpl, ok := registry.Get("iceberg-duckdb")
	require.True(t, ok)
	require.NotNil(t, tmpl.JSONSchema, "iceberg-duckdb should have json_schema")

	result, err := Render(&RenderInput{
		Template:   tmpl,
		Output:     "model",
		DriverSpec: nil, // no driver needed; schema-based
		Properties: map[string]any{
			"aws_access_key_id":     "AKIAEXAMPLE",
			"aws_secret_access_key": "wJalrXUtnFEMI/EXAMPLEKEY",
			"aws_region":            "us-west-2",
			"path":                  "s3://my-iceberg-bucket/warehouse/my_table",
			"name":                  "iceberg_test",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "type: model")
	require.Contains(t, blob, "iceberg_scan")
	require.Contains(t, blob, "s3://my-iceberg-bucket/warehouse/my_table")

	// Verify path uses model_name from "name" property
	require.Equal(t, "models/iceberg_test.yaml", result.Files[0].Path)

	// Verify secret extraction via JSON Schema
	require.Equal(t, "AKIAEXAMPLE", result.EnvVars["AWS_ACCESS_KEY_ID"])
	require.Equal(t, "wJalrXUtnFEMI/EXAMPLEKEY", result.EnvVars["AWS_SECRET_ACCESS_KEY"])

	// Raw secrets should NOT appear in the blob
	require.NotContains(t, blob, "AKIAEXAMPLE")
	require.NotContains(t, blob, "wJalrXUtnFEMI/EXAMPLEKEY")
}

func TestRenderSchemaEnvVarConflict(t *testing.T) {
	tmpl := &Template{
		Name:        "test-schema",
		DisplayName: "Test Schema",
		Tags:        []string{"test"},
		JSONSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"api_key": map[string]any{
					"type":      "string",
					"x-secret":  true,
					"x-env-var": "MY_API_KEY",
				},
			},
		},
		Files: []File{
			{
				Name:         "connector",
				PathTemplate: "connectors/test.yaml",
				CodeTemplate: "type: connector\n[[ renderProps .props ]]",
			},
		},
	}

	// Pre-populate existing env to force conflict
	existingEnv := map[string]bool{"MY_API_KEY": true}

	result, err := Render(&RenderInput{
		Template:    tmpl,
		Properties:  map[string]any{"api_key": "secret123"},
		ExistingEnv: existingEnv,
	})
	require.NoError(t, err)

	// Should use MY_API_KEY_1 due to conflict
	require.Contains(t, result.EnvVars, "MY_API_KEY_1")
	require.Equal(t, "secret123", result.EnvVars["MY_API_KEY_1"])
	require.Contains(t, result.Files[0].Blob, "MY_API_KEY_1")
}

func TestRenderSchemaUIOnlyFieldsSkipped(t *testing.T) {
	tmpl := &Template{
		Name:        "test-ui-only",
		DisplayName: "Test UI Only",
		Tags:        []string{"test"},
		JSONSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"auth_method": map[string]any{
					"type":      "string",
					"x-ui-only": true,
				},
				"host": map[string]any{
					"type": "string",
				},
			},
		},
		Files: []File{
			{
				Name:         "connector",
				PathTemplate: "connectors/test.yaml",
				CodeTemplate: "type: connector\n[[ renderProps .props ]]",
			},
		},
	}

	result, err := Render(&RenderInput{
		Template: tmpl,
		Properties: map[string]any{
			"auth_method": "access_keys",
			"host":        "example.com",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)

	blob := result.Files[0].Blob
	require.Contains(t, blob, "host")
	require.NotContains(t, blob, "auth_method")
}

func TestRenderSchemaPropertyTypes(t *testing.T) {
	tmpl := &Template{
		Name:        "test-types",
		DisplayName: "Test Types",
		Tags:        []string{"test"},
		JSONSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"host": map[string]any{
					"type": "string",
				},
				"port": map[string]any{
					"type": "number",
				},
				"ssl": map[string]any{
					"type": "boolean",
				},
			},
		},
		Files: []File{
			{
				Name:         "connector",
				PathTemplate: "connectors/test.yaml",
				CodeTemplate: "type: connector\n[[ renderProps .props ]]",
			},
		},
	}

	result, err := Render(&RenderInput{
		Template: tmpl,
		Properties: map[string]any{
			"host": "example.com",
			"port": "9440",
			"ssl":  "true",
		},
		ExistingEnv: make(map[string]bool),
	})
	require.NoError(t, err)

	blob := result.Files[0].Blob
	// Strings should be quoted
	require.Contains(t, blob, `host: "example.com"`)
	// Numbers should not be quoted
	require.Contains(t, blob, "port: 9440")
	// Booleans should not be quoted
	require.Contains(t, blob, "ssl: true")
}
