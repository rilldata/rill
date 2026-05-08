import { describe, expect, it, vi, beforeEach } from "vitest";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  replaceAiConnectorInYAML,
  replaceOlapConnectorInYAML,
  replaceOrAddEnvVariable,
  compileConnectorYAML,
  formatHeadersAsYamlMap,
  updateDotEnvWithSecrets,
  maybeUnsetOlapConnectorInYaml,
} from "./code-utils";
import {
  envMappedVarsAndValuesToObject,
  makeTestEnvEditSession,
} from "@rilldata/web-common/features/env-management/test/test-env-store.ts";
import { getGenericEnvVarName } from "@rilldata/web-common/features/connectors/env-utils.ts";
import type { JSONSchemaObject } from "@rilldata/web-common/features/templates/schemas/types.ts";

// Import the template for testing
const YAML_MODEL_TEMPLATE = `type: model
materialize: true\n
connector: {{ connector }}\n
sql: {{ sql }}{{ dev_section }}
`;

describe("YAML Model Template", () => {
  it("should include dev section for non-Redshift connectors", () => {
    const connector = "clickhouse";
    const selectStatement = "select * from my_table";
    const shouldIncludeDevSection = true;
    const devSection = shouldIncludeDevSection
      ? `\ndev:\n  sql: ${selectStatement} limit 10000`
      : "";

    const yamlContent = YAML_MODEL_TEMPLATE.replace(
      "{{ connector }}",
      connector,
    )
      .replace(/{{ sql }}/g, selectStatement)
      .replace("{{ dev_section }}", devSection);

    expect(yamlContent).toContain("dev:");
    expect(yamlContent).toContain("limit 10000");
    expect(yamlContent).toContain("connector: clickhouse");
  });

  it("should exclude dev section for Redshift connector", () => {
    const connector = "redshift";
    const selectStatement = "select * from my_table";
    const shouldIncludeDevSection = false;
    const devSection = shouldIncludeDevSection
      ? `\ndev:\n  sql: ${selectStatement} limit 10000`
      : "";

    const yamlContent = YAML_MODEL_TEMPLATE.replace(
      "{{ connector }}",
      connector,
    )
      .replace(/{{ sql }}/g, selectStatement)
      .replace("{{ dev_section }}", devSection);

    expect(yamlContent).not.toContain("dev:");
    expect(yamlContent).not.toContain("limit 10000");
    expect(yamlContent).toContain("connector: redshift");
    expect(yamlContent).toContain("sql: select * from my_table");
  });
});

describe("replaceOrAddEnvVariable", () => {
  it("should create a new env file", () => {
    const updatedEnvBlob = replaceOrAddEnvVariable("", "KEY1", "VALUE1");
    expect(updatedEnvBlob).toBe("KEY1=VALUE1");
  });

  const existingEnvBlob = `# This is a comment
# This is another comment
KEY1=VALUE1
KEY2=VALUE2`;

  it("should update an existing key in the env file", () => {
    const updatedEnvBlob = replaceOrAddEnvVariable(
      existingEnvBlob,
      "KEY1",
      "NEW_VALUE1",
    );
    expect(updatedEnvBlob).toBe(`# This is a comment
# This is another comment
KEY1=NEW_VALUE1
KEY2=VALUE2`);
  });

  it("should add a new key to the env file", () => {
    const updatedEnvBlob = replaceOrAddEnvVariable(
      existingEnvBlob,
      "KEY3",
      "VALUE3",
    );
    expect(updatedEnvBlob).toBe(`# This is a comment
# This is another comment
KEY1=VALUE1
KEY2=VALUE2
KEY3=VALUE3`);
  });
});

describe("replaceOlapConnectorInYAML", () => {
  it("should add a new `olap_connector` key to a blank file", () => {
    const updatedBlob = replaceOlapConnectorInYAML("", "clickhouse");
    expect(updatedBlob).toBe("olap_connector: clickhouse\n");
  });

  it("should add a new `olap_connector` key to a file with other keys", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n`;
    const updatedBlob = replaceOlapConnectorInYAML(existingBlob, "clickhouse");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nolap_connector: clickhouse\n`,
    );
  });

  it("should update the `olap_connector` key in a file with an existing `olap_connector` key", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n\nolap_connector: snowflake\n`;
    const updatedBlob = replaceOlapConnectorInYAML(existingBlob, "clickhouse");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nolap_connector: clickhouse\n`,
    );
  });
});

describe("maybeUnsetOlapConnectorInYaml", () => {
  it("should not update yaml if olap_connector is not set", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n`;
    const [updated, updatedBlob] = maybeUnsetOlapConnectorInYaml(
      existingBlob,
      "clickhouse",
    );
    expect(updated).toBe(false);
    expect(updatedBlob).toBe(existingBlob);
  });

  it("should unset olap_connector if it is set to the same value", () => {
    const existingBlob = `# here's a comment\ntitle: test project\nolap_connector: clickhouse\nfeatures: ["developer_chat"]`;
    const [updated, updatedBlob] = maybeUnsetOlapConnectorInYaml(
      existingBlob,
      "clickhouse",
    );
    expect(updated).toBe(true);
    expect(updatedBlob).toBe(
      '# here\'s a comment\ntitle: test project\n\nfeatures: ["developer_chat"]',
    );
  });

  it("should not unset olap_connector if it is set to a different value", () => {
    const existingBlob = `# here's a comment\ntitle: test project\nolap_connector: snowflake`;
    const [updated, updatedBlob] = maybeUnsetOlapConnectorInYaml(
      existingBlob,
      "clickhouse",
    );
    expect(updated).toBe(false);
    expect(updatedBlob).toBe(existingBlob);
  });
});

describe("replaceAiConnectorInYAML", () => {
  it("should add a new `ai_connector` key to a blank file", () => {
    const updatedBlob = replaceAiConnectorInYAML("", "claude");
    expect(updatedBlob).toBe("ai_connector: claude\n");
  });

  it("should add a new `ai_connector` key to a file with other keys", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "claude");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nai_connector: claude\n`,
    );
  });

  it("should update the `ai_connector` key in a file with an existing `ai_connector` key", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n\nai_connector: openai\n`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "claude");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nai_connector: claude\n`,
    );
  });

  it("should handle a blob without a trailing newline", () => {
    const existingBlob = `title: test project`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "claude");
    expect(updatedBlob).toBe(`title: test project\nai_connector: claude\n`);
  });

  it("should replace ai_connector in the middle of the file", () => {
    const existingBlob = `title: test project\nai_connector: openai\nolap_connector: clickhouse\n`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "gemini");
    expect(updatedBlob).toBe(
      `title: test project\nai_connector: gemini\nolap_connector: clickhouse\n`,
    );
  });
});

describe("getGenericEnvVarName", () => {
  describe("Schema-driven x-env-var-name (production behavior)", () => {
    it("should use x-env-var-name from schema when provided", () => {
      const schema = {
        properties: {
          key_id: { "x-env-var-name": "GCS_ACCESS_KEY_ID" },
        },
      };
      const result = getGenericEnvVarName("gcs", "key_id", schema);
      expect(result).toBe("GCS_ACCESS_KEY_ID");
    });

    it("should use x-env-var-name for AWS credentials", () => {
      const schema = {
        properties: {
          aws_access_key_id: { "x-env-var-name": "AWS_ACCESS_KEY_ID" },
          aws_secret_access_key: { "x-env-var-name": "AWS_SECRET_ACCESS_KEY" },
        },
      };
      expect(getGenericEnvVarName("s3", "aws_access_key_id", schema)).toBe(
        "AWS_ACCESS_KEY_ID",
      );
      expect(getGenericEnvVarName("s3", "aws_secret_access_key", schema)).toBe(
        "AWS_SECRET_ACCESS_KEY",
      );
    });

    it("should use x-env-var-name for Google credentials", () => {
      const schema = {
        properties: {
          google_application_credentials: {
            "x-env-var-name": "GOOGLE_APPLICATION_CREDENTIALS",
          },
        },
      };
      expect(
        getGenericEnvVarName(
          "bigquery",
          "google_application_credentials",
          schema,
        ),
      ).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should use x-env-var-name for driver-specific secrets", () => {
      const motherduckSchema = {
        properties: {
          token: { "x-env-var-name": "MOTHERDUCK_TOKEN" },
        },
      };
      expect(
        getGenericEnvVarName("motherduck", "token", motherduckSchema),
      ).toBe("MOTHERDUCK_TOKEN");

      const clickhouseSchema = {
        properties: {
          dsn: { "x-env-var-name": "CLICKHOUSE_DSN" },
          password: { "x-env-var-name": "CLICKHOUSE_PASSWORD" },
        },
      };
      expect(getGenericEnvVarName("clickhouse", "dsn", clickhouseSchema)).toBe(
        "CLICKHOUSE_DSN",
      );
      expect(
        getGenericEnvVarName("clickhouse", "password", clickhouseSchema),
      ).toBe("CLICKHOUSE_PASSWORD");
    });
  });

  describe("Fallback behavior (DRIVERNAME_PROPERTYKEY format)", () => {
    it("should use DRIVERNAME_PROPERTYKEY when schema has no x-env-var-name", () => {
      const schema = {
        properties: {
          other_field: { "x-env-var-name": "OTHER_VAR" },
        },
      };
      const result = getGenericEnvVarName("gcs", "key_id", schema);
      expect(result).toBe("GCS_KEY_ID");
    });

    it("should use DRIVERNAME_PROPERTYKEY when schema is undefined", () => {
      const result = getGenericEnvVarName("motherduck", "token", undefined);
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should use DRIVERNAME_PROPERTYKEY when schema has no properties", () => {
      const schema = { properties: {} };
      const result = getGenericEnvVarName("motherduck", "token", schema);
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should convert camelCase to SCREAMING_SNAKE_CASE", () => {
      const result = getGenericEnvVarName("bigquery", "projectId");
      expect(result).toBe("BIGQUERY_PROJECT_ID");
    });

    it("should handle dots and hyphens by converting to underscores", () => {
      const result = getGenericEnvVarName(
        "custom",
        "property.with-mixed.separators",
      );
      expect(result).toBe("CUSTOM_PROPERTY_WITH_MIXED_SEPARATORS");
    });
  });
});

describe("formatHeadersAsYamlMap", () => {
  describe("array input", () => {
    it("should format non-sensitive headers as plain text", () => {
      const result = formatHeadersAsYamlMap([
        { key: "Content-Type", value: "application/json" },
        { key: "Accept", value: "text/html" },
      ]);
      expect(result).toBe(
        `headers:\n    "Content-Type": "application/json"\n    "Accept": "text/html"`,
      );
    });

    it("should replace sensitive header with env ref when driverName provided", () => {
      const result = formatHeadersAsYamlMap(
        [{ key: "Authorization", value: "my_secret_token" }],
        "https",
      );
      expect(result).toContain(
        '"Authorization": "{{ .env.connector.https.authorization }}"',
      );
    });

    it("should preserve Bearer scheme prefix", () => {
      const result = formatHeadersAsYamlMap(
        [{ key: "Authorization", value: "Bearer my_token" }],
        "https",
      );
      expect(result).toContain(
        '"Authorization": "Bearer {{ .env.connector.https.authorization }}"',
      );
    });

    it("should preserve Basic scheme prefix", () => {
      const result = formatHeadersAsYamlMap(
        [{ key: "Authorization", value: "Basic dXNlcjpwYXNz" }],
        "https",
      );
      expect(result).toContain(
        '"Authorization": "Basic {{ .env.connector.https.authorization }}"',
      );
    });

    it("should handle mixed sensitive and non-sensitive headers", () => {
      const result = formatHeadersAsYamlMap(
        [
          { key: "Content-Type", value: "application/json" },
          { key: "Authorization", value: "Bearer token123" },
        ],
        "https",
      );
      expect(result).toContain('"Content-Type": "application/json"');
      expect(result).toContain(
        '"Authorization": "Bearer {{ .env.connector.https.authorization }}"',
      );
    });

    it("should filter entries with empty keys", () => {
      const result = formatHeadersAsYamlMap([
        { key: "", value: "ignored" },
        { key: "Accept", value: "text/html" },
      ]);
      expect(result).toBe(`headers:\n    "Accept": "text/html"`);
    });

    it("should return empty string for empty array", () => {
      expect(formatHeadersAsYamlMap([])).toBe("");
    });

    it("should use connectorInstanceName for env refs when provided", () => {
      const result = formatHeadersAsYamlMap(
        [{ key: "X-API-Key", value: "secret" }],
        "https",
        "my_api",
      );
      expect(result).toContain("{{ .env.connector.my_api.x_api_key }}");
    });

    it("should not create env refs when no driverName", () => {
      const result = formatHeadersAsYamlMap([
        { key: "Authorization", value: "Bearer token" },
      ]);
      expect(result).toContain('"Authorization": "Bearer token"');
      expect(result).not.toContain(".env.");
    });
  });

  describe("string input (legacy)", () => {
    it("should parse Key: Value lines", () => {
      const result = formatHeadersAsYamlMap(
        "Content-Type: application/json\nAccept: text/html",
      );
      expect(result).toBe(
        `headers:\n    "Content-Type": "application/json"\n    "Accept": "text/html"`,
      );
    });

    it("should replace sensitive headers with env refs", () => {
      const result = formatHeadersAsYamlMap(
        "Authorization: Bearer my_token",
        "https",
      );
      expect(result).toContain(
        "Bearer {{ .env.connector.https.authorization }}",
      );
    });

    it("should return empty string for empty input", () => {
      expect(formatHeadersAsYamlMap("")).toBe("");
    });
  });
});

describe("compileConnectorYAML", () => {
  it("should produce basic connector YAML", async () => {
    const connector: V1ConnectorDriver = {
      name: "clickhouse",
      docsUrl:
        "https://docs.rilldata.com/developers/build/connectors/data-source/clickhouse",
    };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com" },
      envEditSession,
      {
        orderedProperties: [{ key: "host" }],
      },
    );
    expect(result).toContain("# Connector YAML");
    expect(result).toContain("type: connector");
    expect(result).toContain("driver: clickhouse");
    expect(result).toContain("host: ch.example.com");
  });

  it("should preserve property ordering from orderedProperties", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com", port: 9000, database: "default" },
      envEditSession,
      {
        orderedProperties: [
          { key: "database" },
          { key: "host" },
          { key: "port" },
        ],
      },
    );
    const dbIdx = result.indexOf("database:");
    const hostIdx = result.indexOf("host:");
    const portIdx = result.indexOf("port:");
    expect(dbIdx).toBeLessThan(hostIdx);
    expect(hostIdx).toBeLessThan(portIdx);
  });

  it("should replace secret properties with env var placeholders", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const schema = {
      type: "object",
      properties: {
        password: { "x-env-var-name": "CLICKHOUSE_PASSWORD" },
      },
    } satisfies JSONSchemaObject;
    const { testEnvs, envEditSession } = await makeTestEnvEditSession(
      connector.name,
      schema,
    );
    const result = compileConnectorYAML(
      connector,
      { password: "super_secret" },
      envEditSession,
      {
        orderedProperties: [{ key: "password", secret: true }],
        secretKeys: ["password"],
        schema,
      },
    );
    expect(result).toContain("{{ .env.CLICKHOUSE_PASSWORD }}");
    expect(result).not.toContain("super_secret");
    expect(envEditSession.entries.get("password")?.mappedEnvVarName).toEqual(
      "CLICKHOUSE_PASSWORD",
    );

    // Value is saved in env only after a flush
    expect(testEnvs).toEqual({});
    await envEditSession.commit();
    expect(testEnvs).toEqual({
      CLICKHOUSE_PASSWORD: "super_secret",
    });
  });

  it("should handle env var conflict resolution with env edit session", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const schema = {
      type: "object",
      properties: {
        password: { "x-env-var-name": "CLICKHOUSE_PASSWORD" },
      },
    } satisfies JSONSchemaObject;
    const { testEnvs, envEditSession } = await makeTestEnvEditSession(
      connector.name,
      schema,
      {},
      {
        CLICKHOUSE_PASSWORD: "abc",
      },
    );
    const result = compileConnectorYAML(
      connector,
      { password: "secret" },
      envEditSession,
      {
        orderedProperties: [{ key: "password", secret: true }],
        secretKeys: ["password"],
        schema,
      },
    );
    expect(result).toContain("CLICKHOUSE_PASSWORD_1");

    // Value is saved in env only after a flush
    expect(testEnvs).toEqual({
      CLICKHOUSE_PASSWORD: "abc",
    });
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      CLICKHOUSE_PASSWORD_1: "secret",
    });

    // Calling compile again should not create new variable.
    const newResult = compileConnectorYAML(
      connector,
      { password: "secret_new" },
      envEditSession,
      {
        orderedProperties: [{ key: "password", secret: true }],
        secretKeys: ["password"],
        schema,
      },
    );
    expect(newResult).toContain("CLICKHOUSE_PASSWORD_1");

    expect(testEnvs).toEqual({
      CLICKHOUSE_PASSWORD: "abc",
    });
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      CLICKHOUSE_PASSWORD_1: "secret_new",
    });
    await envEditSession.commit();
    expect(testEnvs).toEqual({
      CLICKHOUSE_PASSWORD: "abc",
      CLICKHOUSE_PASSWORD_1: "secret_new",
    });
  });

  it("should quote string properties", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com" },
      envEditSession,
      {
        orderedProperties: [{ key: "host" }],
        stringKeys: ["host"],
      },
    );
    expect(result).toContain('host: "ch.example.com"');
  });

  it("should not quote non-string properties", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { port: 9000 },
      envEditSession,
      { orderedProperties: [{ key: "port" }] },
    );
    expect(result).toContain("port: 9000");
    expect(result).not.toContain('"9000"');
  });

  it("should filter out empty string values", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com", database: "" },
      envEditSession,
      { orderedProperties: [{ key: "host" }, { key: "database" }] },
    );
    expect(result).toContain("host:");
    expect(result).not.toContain("database:");
  });

  it("should filter out undefined values", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com", database: undefined },
      envEditSession,
      { orderedProperties: [{ key: "host" }, { key: "database" }] },
    );
    expect(result).not.toContain("database:");
  });

  it("should filter out empty arrays", async () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { url: "https://example.com", headers: [] },
      envEditSession,
      { orderedProperties: [{ key: "url" }, { key: "headers" }] },
    );
    expect(result).not.toContain("headers:");
  });

  it("should exclude clickhouse managed: false", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com", managed: false },
      envEditSession,
      { orderedProperties: [{ key: "host" }, { key: "managed" }] },
    );
    expect(result).not.toContain("managed");
  });

  it("should include clickhouse managed: true", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com", managed: true },
      envEditSession,
      { orderedProperties: [{ key: "host" }, { key: "managed" }] },
    );
    expect(result).toContain("managed: true");
  });

  it("should output driver as duckdb for motherduck", async () => {
    const connector: V1ConnectorDriver = { name: "motherduck" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { path: "md:my_db" },
      envEditSession,
      { orderedProperties: [{ key: "path" }] },
    );
    expect(result).toContain("driver: duckdb");
    expect(result).not.toContain("driver: motherduck");
  });

  it("should apply fieldFilter to exclude internal properties", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com", managed: true },
      envEditSession,
      {
        orderedProperties: [
          { key: "host" },
          { key: "managed", internal: true },
        ],
        fieldFilter: (p) => !("internal" in p && p.internal),
      },
    );
    expect(result).toContain("host:");
    expect(result).not.toContain("managed:");
  });

  it("should produce no property lines when orderedProperties is empty", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = compileConnectorYAML(
      connector,
      { host: "ch.example.com" },
      envEditSession,
    );
    expect(result).toContain("type: connector");
    expect(result).toContain("driver: clickhouse");
    expect(result).not.toContain("host:");
  });
});

describe("updateDotEnvWithSecrets", () => {
  const mockClient = {
    instanceId: "test-instance-id",
  } as unknown as RuntimeClient;

  // Track fetchQuery calls so tests can inspect them
  let mockEnvBlob = "";
  const mockQueryClient = {
    invalidateQueries: vi.fn().mockResolvedValue(undefined),
    fetchQuery: vi
      .fn()
      .mockImplementation(() => Promise.resolve({ blob: mockEnvBlob })),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockEnvBlob = "";
    mockQueryClient.fetchQuery.mockImplementation(() =>
      Promise.resolve({ blob: mockEnvBlob }),
    );
  });

  it("should add secret keys to empty .env", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = {
      password: "my_secret",
      sql: "SELECT 1",
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["password"] },
    );
    expect(newBlob).toContain("CLICKHOUSE_PASSWORD=my_secret");
  });

  it("should add multiple secret keys", async () => {
    const connector: V1ConnectorDriver = { name: "s3" };
    const formValues: Record<string, unknown> = {
      aws_access_key_id: "AKID123",
      aws_secret_access_key: "SECRET456",
    };
    const schema = {
      properties: {
        aws_access_key_id: { "x-env-var-name": "AWS_ACCESS_KEY_ID" },
        aws_secret_access_key: {
          "x-env-var-name": "AWS_SECRET_ACCESS_KEY",
        },
      },
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["aws_access_key_id", "aws_secret_access_key"], schema },
    );
    expect(newBlob).toContain("AWS_ACCESS_KEY_ID=AKID123");
    expect(newBlob).toContain("AWS_SECRET_ACCESS_KEY=SECRET456");
  });

  it("should append to existing .env without overwriting", async () => {
    mockEnvBlob = "EXISTING_VAR=existing_value";
    mockQueryClient.fetchQuery.mockResolvedValue({ blob: mockEnvBlob });

    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = { password: "new_pw" };
    const { newBlob, originalBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["password"] },
    );
    expect(originalBlob).toBe("EXISTING_VAR=existing_value");
    expect(newBlob).toContain("EXISTING_VAR=existing_value");
    expect(newBlob).toContain("CLICKHOUSE_PASSWORD=new_pw");
  });

  it("should handle env var conflicts with _1 suffix", async () => {
    mockEnvBlob = "CLICKHOUSE_PASSWORD=old_value";
    mockQueryClient.fetchQuery.mockResolvedValue({ blob: mockEnvBlob });

    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = { password: "new_value" };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["password"] },
    );
    // Should use _1 suffix since base name already exists
    expect(newBlob).toContain("CLICKHOUSE_PASSWORD=old_value");
    expect(newBlob).toContain("CLICKHOUSE_PASSWORD_1=new_value");
  });

  it("should skip empty or missing secret values", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = {
      password: "",
      dsn: undefined,
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["password", "dsn"] },
    );
    expect(newBlob).toBe("");
  });

  it("should persist sensitive header values as env entries", async () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      headers: [
        { key: "Authorization", value: "Bearer my_token" },
        { key: "Content-Type", value: "application/json" },
      ],
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: [] },
    );
    // Authorization is sensitive — secret part stored (without Bearer prefix)
    expect(newBlob).toContain("my_token");
    // Content-Type is not sensitive — should NOT be in .env
    expect(newBlob).not.toContain("application/json");
  });

  it("should extract secret from Bearer scheme for sensitive headers", async () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      headers: [{ key: "Authorization", value: "Bearer abc123" }],
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: [] },
    );
    // Only the secret portion (after "Bearer ") is stored
    expect(newBlob).toContain("=abc123");
    expect(newBlob).not.toContain("Bearer");
  });

  it("should store full value when no auth scheme prefix", async () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      headers: [{ key: "X-API-Key", value: "raw_api_key_value" }],
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: [] },
    );
    expect(newBlob).toContain("=raw_api_key_value");
  });

  it("should handle both secrets and sensitive headers together", async () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      password: "http_pass",
      headers: [{ key: "Authorization", value: "Token secret_tok" }],
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["password"] },
    );
    expect(newBlob).toContain("HTTPS_PASSWORD=http_pass");
    expect(newBlob).toContain("secret_tok");
  });

  it("should skip headers with empty keys or values", async () => {
    const connector: V1ConnectorDriver = { name: "https" };
    const formValues: Record<string, unknown> = {
      headers: [
        { key: "", value: "some_value" },
        { key: "Authorization", value: "" },
        { key: "  ", value: "Bearer token" },
      ],
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: [] },
    );
    expect(newBlob).toBe("");
  });

  it("should invalidate cache before reading .env", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      { password: "pw" },
      { secretKeys: ["password"] },
    );
    expect(mockQueryClient.invalidateQueries).toHaveBeenCalledTimes(1);
    // invalidateQueries should be called before fetchQuery
    const invalidateOrder =
      mockQueryClient.invalidateQueries.mock.invocationCallOrder[0];
    const fetchOrder = mockQueryClient.fetchQuery.mock.invocationCallOrder[0];
    expect(invalidateOrder).toBeLessThan(fetchOrder);
  });

  it("should handle missing .env file gracefully", async () => {
    mockQueryClient.fetchQuery.mockRejectedValue({
      response: { data: { message: "no such file" } },
    });

    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { newBlob, originalBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      { password: "pw" },
      { secretKeys: ["password"] },
    );
    expect(originalBlob).toBe("");
    expect(newBlob).toContain("CLICKHOUSE_PASSWORD=pw");
  });

  it("should rethrow non-file-not-found errors", async () => {
    mockQueryClient.fetchQuery.mockRejectedValue({
      response: { data: { message: "permission denied" } },
    });

    const connector: V1ConnectorDriver = { name: "clickhouse" };
    await expect(
      updateDotEnvWithSecrets(
        mockClient,
        mockQueryClient as any,
        connector,
        { password: "pw" },
        { secretKeys: ["password"] },
      ),
    ).rejects.toEqual({
      response: { data: { message: "permission denied" } },
    });
  });

  it("should use originalBlob for conflict detection across all secrets", async () => {
    // When adding multiple secrets, conflict detection should use the original blob,
    // not the progressively updated one
    mockEnvBlob = "";
    mockQueryClient.fetchQuery.mockResolvedValue({ blob: mockEnvBlob });

    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const formValues: Record<string, unknown> = {
      password: "pw1",
      dsn: "clickhouse://...",
    };
    const { newBlob } = await updateDotEnvWithSecrets(
      mockClient,
      mockQueryClient as any,
      connector,
      formValues,
      { secretKeys: ["password", "dsn"] },
    );
    // Both should use base name since originalBlob is empty
    expect(newBlob).toContain("CLICKHOUSE_PASSWORD=pw1");
    expect(newBlob).toContain("CLICKHOUSE_DSN=clickhouse://...");
  });
});
