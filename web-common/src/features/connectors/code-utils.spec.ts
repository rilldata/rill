import { describe, expect, it } from "vitest";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import {
  replaceAiConnectorInYAML,
  replaceOlapConnectorInYAML,
  generateYAML,
  formatHeadersAsYamlMap,
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
    it("should format non-sensitive headers as plain text", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        [
          { key: "Content-Type", value: "application/json" },
          { key: "Accept", value: "text/html" },
        ],
        envEditSession,
      );
      expect(result).toBe(
        `headers:\n    "Content-Type": "application/json"\n    "Accept": "text/html"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({});
    });

    it("should replace sensitive header with env ref when driverName provided", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        [{ key: "Authorization", value: "my_secret_token" }],
        envEditSession,
      );
      expect(result).toContain(
        '"Authorization": "{{ .env.HTTPS_AUTHORIZATION }}"',
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        HTTPS_AUTHORIZATION: "my_secret_token",
      });
    });

    it("should preserve Bearer scheme prefix", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        [{ key: "Authorization", value: "Bearer my_token" }],
        envEditSession,
      );
      expect(result).toContain(
        '"Authorization": "Bearer {{ .env.HTTPS_AUTHORIZATION }}"',
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        HTTPS_AUTHORIZATION: "my_token",
      });
    });

    it("should handle mixed sensitive and non-sensitive headers", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        [
          { key: "Content-Type", value: "application/json" },
          { key: "Authorization", value: "Bearer token123" },
        ],
        envEditSession,
      );
      expect(result).toContain('"Content-Type": "application/json"');
      expect(result).toContain(
        '"Authorization": "Bearer {{ .env.HTTPS_AUTHORIZATION }}"',
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        HTTPS_AUTHORIZATION: "token123",
      });
    });

    it("should filter entries with empty keys", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        [
          { key: "", value: "ignored" },
          { key: "Accept", value: "text/html" },
        ],
        envEditSession,
      );
      expect(result).toBe(`headers:\n    "Accept": "text/html"`);
      await envEditSession.commit();
      expect(testEnvs).toEqual({});
    });

    it("should return empty string for empty array", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      expect(formatHeadersAsYamlMap([], envEditSession)).toBe("");
      await envEditSession.commit();
      expect(testEnvs).toEqual({});
    });
  });

  describe("string input (legacy)", () => {
    it("should parse Key: Value lines", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        "Content-Type: application/json\nAccept: text/html",
        envEditSession,
      );
      expect(result).toBe(
        `headers:\n    "Content-Type": "application/json"\n    "Accept": "text/html"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({});
    });

    it("should replace sensitive headers with env refs", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      const result = formatHeadersAsYamlMap(
        "Authorization: Bearer my_token",
        envEditSession,
      );
      expect(result).toContain("Bearer {{ .env.HTTPS_AUTHORIZATION }}");
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        HTTPS_AUTHORIZATION: "my_token",
      });
    });

    it("should return empty string for empty input", async () => {
      const { testEnvs, envEditSession } = await makeTestEnvEditSession(
        "https",
        undefined,
      );
      expect(formatHeadersAsYamlMap("", envEditSession)).toBe("");
      await envEditSession.commit();
      expect(testEnvs).toEqual({});
    });
  });
});

describe("generateYAML", () => {
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const newResult = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(connector, { port: 9000 }, envEditSession, {
      orderedProperties: [{ key: "port" }],
    });
    expect(result).toContain("port: 9000");
    expect(result).not.toContain('"9000"');
  });

  it("should filter out empty string values", async () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const { envEditSession } = await makeTestEnvEditSession(
      connector.name,
      undefined,
    );
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
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
    const result = generateYAML(
      connector,
      { host: "ch.example.com" },
      envEditSession,
    );
    expect(result).toContain("type: connector");
    expect(result).toContain("driver: clickhouse");
    expect(result).not.toContain("host:");
  });
});
