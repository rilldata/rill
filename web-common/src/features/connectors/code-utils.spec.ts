import { describe, expect, it } from "vitest";
import {
  replaceOlapConnectorInYAML,
  replaceOrAddEnvVariable,
} from "./code-utils";

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
