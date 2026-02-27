import { describe, it, expect } from "vitest";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import {
  getFileTypeFromPath,
  inferSourceName,
  buildDuckDbQuery,
  maybeRewriteToDuckDb,
  compileSourceYAML,
  prepareSourceFormData,
} from "./sourceUtils";

const gcsTests = [
  {
    path: "gs://bucket-name/folder-name/file-name.csv",
    expected: "file_name",
    expectedExtension: "csv",
  },
  {
    path: "gs://bucket-name/folder-name/file-name.parquet",
    expected: "file_name",
    expectedExtension: "parquet",
  },
  {
    path: "gs://bucket-name/folder-name/file-name.csv.gz",
    expected: "file_name",
    expectedExtension: "csv",
  },
  {
    path: "gs://bucket-name/folder-name/file-name",
    expected: "file_name",
    expectedExtension: "",
  },
  {
    path: "gs://bucket-name/folder-name/FILE-NAME.csv",
    expected: "FILE_NAME",
    expectedExtension: "csv",
  },
  {
    path: "gs://bucket-name/folder-name/file-name123",
    expected: "file_name123",
    expectedExtension: "",
  },
];

const s3Tests = [
  {
    path: "s3://bucket-name/folder-name/file-name.csv",
    expected: "file_name",
  },
];

const httpTests = [
  {
    path: "http://example.com/folder-name/file-name.csv",
    expected: "file_name",
  },
  {
    path: "https://example.com/folder-name/file-name.csv",
    expected: "file_name",
  },
];

describe("inferSourceName", () => {
  // GCS
  it("should infer source name for GCS connector", () => {
    const connector = {
      name: "gcs",
    };

    gcsTests.forEach((test) => {
      const actual = inferSourceName(connector, test.path);
      expect(actual).toEqual(test.expected);
    });
  });

  // S3
  it("should infer source name for S3 connector", () => {
    const connector = {
      name: "s3",
    };

    s3Tests.forEach((test) => {
      const actual = inferSourceName(connector, test.path);
      expect(actual).toEqual(test.expected);
    });
  });

  // HTTPS
  it("should infer source name for HTTPS connector", () => {
    const connector = {
      name: "https",
    };

    httpTests.forEach((test) => {
      const actual = inferSourceName(connector, test.path);
      expect(actual).toEqual(test.expected);
    });
  });
});

describe("getFileTypeFromPath", () => {
  it("should infer source name for given path", () => {
    gcsTests.forEach((test) => {
      const actual = getFileTypeFromPath(test.path);
      expect(actual).toEqual(test.expectedExtension);
    });
  });
});

describe("buildDuckDbQuery", () => {
  const cases: Array<{
    name: string;
    path: string | undefined;
    options?: { defaultToJson?: boolean };
    expected: string;
  }> = [
    {
      name: "csv",
      path: "s3://bucket/data.csv",
      expected:
        "select * from read_csv('s3://bucket/data.csv', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "tsv",
      path: "s3://bucket/data.tsv",
      expected:
        "select * from read_csv('s3://bucket/data.tsv', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "txt",
      path: "/data/file.txt",
      expected:
        "select * from read_csv('/data/file.txt', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "parquet",
      path: "s3://bucket/data.parquet",
      expected: "select * from read_parquet('s3://bucket/data.parquet')",
    },
    {
      name: "json",
      path: "gs://bucket/data.json",
      expected:
        "select * from read_json('gs://bucket/data.json', auto_detect=true, format='auto')",
    },
    {
      name: "ndjson",
      path: "gs://bucket/data.ndjson",
      expected:
        "select * from read_json('gs://bucket/data.ndjson', auto_detect=true, format='auto')",
    },
    {
      name: "compound extension .v1.parquet.gz",
      path: "s3://bucket/data.v1.parquet.gz",
      expected: "select * from read_parquet('s3://bucket/data.v1.parquet.gz')",
    },
    {
      name: "compound extension .csv.gz",
      path: "s3://bucket/data.csv.gz",
      expected:
        "select * from read_csv('s3://bucket/data.csv.gz', auto_detect=true, ignore_errors=1, header=true)",
    },
    {
      name: "unknown extension without defaultToJson",
      path: "s3://bucket/data.avro",
      expected: "select * from 's3://bucket/data.avro'",
    },
    {
      name: "unknown extension with defaultToJson",
      path: "https://api.example.com/data",
      options: { defaultToJson: true },
      expected:
        "select * from read_json('https://api.example.com/data', auto_detect=true, format='auto')",
    },
    {
      name: "undefined path",
      path: undefined,
      expected: "select * from ''",
    },
  ];

  for (const tc of cases) {
    it(tc.name, () => {
      expect(buildDuckDbQuery(tc.path, tc.options)).toBe(tc.expected);
    });
  }
});

describe("maybeRewriteToDuckDb", () => {
  it("should rewrite s3 with connector name", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      path: "s3://bucket/data.parquet",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues, {
      connectorInstanceName: "my_s3",
    });
    expect(result.name).toBe("duckdb");
    expect(values.sql).toBe(
      "select * from read_parquet('s3://bucket/data.parquet')",
    );
    expect(values.path).toBeUndefined();
    expect(values.create_secrets_from_connectors).toBe("my_s3");
  });

  it("should rewrite s3 without connector name — uses driver name for secrets", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      path: "s3://bucket/data.csv",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues);
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_csv");
    expect(values.create_secrets_from_connectors).toBe("s3");
  });

  it("should rewrite gcs", () => {
    const connector: V1ConnectorDriver = { name: "gcs" };
    const formValues: Record<string, unknown> = {
      path: "gs://bucket/data.json",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues, {
      connectorInstanceName: "my_gcs",
    });
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_json");
    expect(values.create_secrets_from_connectors).toBe("my_gcs");
  });

  it("should rewrite azure", () => {
    const connector: V1ConnectorDriver = { name: "azure" };
    const formValues: Record<string, unknown> = {
      path: "azure://container/data.parquet",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues, {
      connectorInstanceName: "my_azure",
    });
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_parquet");
    expect(values.create_secrets_from_connectors).toBe("my_azure");
  });

  it("should rewrite https with connector name", () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      path: "https://api.example.com/data",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues, {
      connectorInstanceName: "my_http",
    });
    expect(result.name).toBe("duckdb");
    // https defaults to read_json
    expect(values.sql).toContain("read_json");
    expect(values.create_secrets_from_connectors).toBe("my_http");
  });

  it("should rewrite https without connector name — no create_secrets", () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      path: "https://example.com/data.csv",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues);
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_csv");
    expect(values.create_secrets_from_connectors).toBeUndefined();
  });

  it("should rewrite local_file", () => {
    const connector: V1ConnectorDriver = { name: "local_file" };
    const formValues: Record<string, unknown> = { path: "/data/file.parquet" };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues);
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_parquet");
    expect(values.path).toBeUndefined();
  });

  it("should rewrite sqlite with sqlite_scan", () => {
    const connector: V1ConnectorDriver = { name: "sqlite" };
    const formValues: Record<string, unknown> = {
      db: "/data/app.db",
      table: "users",
    };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues);
    expect(result.name).toBe("duckdb");
    expect(values.sql).toBe(
      "SELECT * FROM sqlite_scan('/data/app.db', 'users');",
    );
    expect(values.db).toBeUndefined();
    expect(values.table).toBeUndefined();
  });

  it("should not rewrite clickhouse", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = { sql: "SELECT 1" };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues);
    expect(result.name).toBe("clickhouse");
    expect(values.sql).toBe("SELECT 1");
  });

  it("should not rewrite postgres", () => {
    const connector: V1ConnectorDriver = { name: "postgres" };
    const formValues: Record<string, unknown> = { sql: "SELECT 1" };
    const [result, values] = maybeRewriteToDuckDb(connector, formValues);
    expect(result.name).toBe("postgres");
    expect(values.sql).toBe("SELECT 1");
  });

  it("should not mutate the original connector object", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      path: "s3://bucket/data.csv",
    };
    maybeRewriteToDuckDb(connector, formValues);
    expect(connector.name).toBe("s3");
  });

  it("should preserve existing create_secrets_from_connectors for s3", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      path: "s3://bucket/data.csv",
      create_secrets_from_connectors: "existing_connector",
    };
    const [, values] = maybeRewriteToDuckDb(connector, formValues, {
      connectorInstanceName: "new_s3",
    });
    // Should preserve existing value, not overwrite
    expect(values.create_secrets_from_connectors).toBe("existing_connector");
  });
});

describe("compileSourceYAML", () => {
  it("should produce basic model YAML with SQL", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(connector, {
      sql: "SELECT * FROM events",
    });
    expect(result).toContain("# Model YAML");
    expect(result).toContain("type: model");
    expect(result).toContain("materialize: true");
    expect(result).toContain("connector: clickhouse");
    expect(result).toContain("sql: |");
    expect(result).toContain("  SELECT * FROM events");
  });

  it("should replace secret properties with env var placeholders", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(
      connector,
      { password: "super_secret", sql: "SELECT 1" },
      { secretKeys: ["password"] },
    );
    expect(result).toContain("{{ .env.CLICKHOUSE_PASSWORD }}");
    expect(result).not.toContain("super_secret");
  });

  it("should quote string properties", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(
      connector,
      { host: "ch.example.com", sql: "SELECT 1" },
      { stringKeys: ["host"] },
    );
    expect(result).toContain('host: "ch.example.com"');
  });

  it("should not quote non-string properties", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(connector, {
      port: 9000,
      sql: "SELECT 1",
    });
    expect(result).toContain("port: 9000");
    expect(result).not.toContain('port: "9000"');
  });

  it("should filter out empty string values", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(connector, {
      database: "",
      sql: "SELECT 1",
    });
    expect(result).not.toContain("database:");
  });

  it("should filter out undefined values", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(connector, {
      database: undefined,
      sql: "SELECT 1",
    });
    expect(result).not.toContain("database:");
  });

  it("should always exclude the name field", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(connector, {
      name: "my_source",
      sql: "SELECT 1",
    });
    expect(result).not.toContain("name: my_source");
  });

  it("should include dev section for warehouse connectors", () => {
    const connector: V1ConnectorDriver = {
      name: "clickhouse",
      implementsWarehouse: true,
    };
    const result = compileSourceYAML(connector, {
      sql: "SELECT * FROM events;",
    });
    expect(result).toContain("dev:");
    expect(result).toContain("limit 10000");
    // Dev SQL should strip trailing semicolons
    expect(result).toContain("SELECT * FROM events limit 10000");
  });

  it("should skip dev section for redshift", () => {
    const connector: V1ConnectorDriver = {
      name: "redshift",
      implementsWarehouse: true,
    };
    const result = compileSourceYAML(connector, {
      sql: "SELECT * FROM events",
    });
    expect(result).not.toContain("dev:");
  });

  it("should skip dev section for non-warehouse connectors", () => {
    const connector: V1ConnectorDriver = {
      name: "duckdb",
      implementsWarehouse: false,
    };
    const result = compileSourceYAML(connector, {
      sql: "SELECT * FROM events",
    });
    expect(result).not.toContain("dev:");
  });

  it("should skip dev section when no SQL", () => {
    const connector: V1ConnectorDriver = {
      name: "clickhouse",
      implementsWarehouse: true,
    };
    const result = compileSourceYAML(connector, { host: "ch.example.com" });
    expect(result).not.toContain("dev:");
  });

  it("should use connectorInstanceName as connector value", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(
      connector,
      { sql: "SELECT 1" },
      { connectorInstanceName: "clickhouse_prod" },
    );
    expect(result).toContain("connector: clickhouse_prod");
  });

  it("should use originalDriverName in header comment", () => {
    const connector: V1ConnectorDriver = { name: "duckdb" };
    const result = compileSourceYAML(
      connector,
      { sql: "SELECT 1" },
      { originalDriverName: "s3" },
    );
    expect(result).toContain(
      "https://docs.rilldata.com/developers/build/connectors/data-source/s3",
    );
  });

  it("should handle env var conflict resolution with existingEnvBlob", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const result = compileSourceYAML(
      connector,
      { password: "secret", sql: "SELECT 1" },
      {
        secretKeys: ["password"],
        existingEnvBlob: "CLICKHOUSE_PASSWORD=old_value",
      },
    );
    expect(result).toContain("CLICKHOUSE_PASSWORD_1");
  });
});

describe("prepareSourceFormData", () => {
  it("should strip auth_method from form values", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      auth_method: "access_keys",
      path: "s3://bucket/data.parquet",
    };
    const [, values] = prepareSourceFormData(connector, formValues);
    expect(values.auth_method).toBeUndefined();
  });

  it("should strip connector-step fields for s3", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      aws_access_key_id: "AKID",
      aws_secret_access_key: "secret",
      region: "us-east-1",
      endpoint: "https://s3.example.com",
      path: "s3://bucket/data.csv",
    };
    const [, values] = prepareSourceFormData(connector, formValues);
    // Connector-level fields should be removed
    expect(values.aws_access_key_id).toBeUndefined();
    expect(values.aws_secret_access_key).toBeUndefined();
    expect(values.region).toBeUndefined();
    expect(values.endpoint).toBeUndefined();
  });

  it("should apply DuckDB rewrite for s3", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      path: "s3://bucket/data.parquet",
    };
    const [result, values] = prepareSourceFormData(connector, formValues, {
      connectorInstanceName: "my_s3",
    });
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_parquet");
    expect(values.path).toBeUndefined();
  });

  it("should not mutate original formValues", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      auth_method: "access_keys",
      aws_access_key_id: "AKID",
      path: "s3://bucket/data.csv",
    };
    prepareSourceFormData(connector, formValues);
    // Original should be preserved
    expect(formValues.auth_method).toBe("access_keys");
    expect(formValues.aws_access_key_id).toBe("AKID");
    expect(formValues.path).toBe("s3://bucket/data.csv");
  });

  it("should not mutate original connector", () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    prepareSourceFormData(connector, { path: "s3://bucket/data.csv" });
    expect(connector.name).toBe("s3");
  });

  it("should pass through non-rewritten connectors", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = {
      sql: "SELECT * FROM events",
    };
    const [result, values] = prepareSourceFormData(connector, formValues);
    expect(result.name).toBe("clickhouse");
    expect(values.sql).toBe("SELECT * FROM events");
  });

  it("should strip connector-step fields for clickhouse", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = {
      host: "ch.example.com",
      port: 8443,
      username: "default",
      password: "secret",
      sql: "SELECT * FROM events",
    };
    const [, values] = prepareSourceFormData(connector, formValues);
    // All connector-step fields removed
    expect(values.host).toBeUndefined();
    expect(values.port).toBeUndefined();
    expect(values.username).toBeUndefined();
    expect(values.password).toBeUndefined();
    // Source-step fields preserved
    expect(values.sql).toBe("SELECT * FROM events");
  });

  it("should handle https with DuckDB rewrite and defaultToJson", () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      path: "https://api.example.com/data",
    };
    const [result, values] = prepareSourceFormData(connector, formValues);
    expect(result.name).toBe("duckdb");
    expect(values.sql).toContain("read_json");
  });

  it("should handle connector with no schema gracefully", () => {
    const connector: V1ConnectorDriver = { name: "unknown_connector" };
    const formValues: Record<string, unknown> = {
      sql: "SELECT 1",
      some_field: "value",
    };
    const [result, values] = prepareSourceFormData(connector, formValues);
    // No schema, so no stripping/placeholder logic applies
    expect(result.name).toBe("unknown_connector");
    expect(values.sql).toBe("SELECT 1");
    expect(values.some_field).toBe("value");
  });

  it("should strip auth_method even when no schema exists", () => {
    const connector: V1ConnectorDriver = { name: "unknown_connector" };
    const formValues: Record<string, unknown> = {
      auth_method: "oauth",
      sql: "SELECT 1",
    };
    const [, values] = prepareSourceFormData(connector, formValues);
    expect(values.auth_method).toBeUndefined();
  });
});
