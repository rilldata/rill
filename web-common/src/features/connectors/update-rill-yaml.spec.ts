import { describe, expect, it } from "vitest";
import { updateRillYAMLBlobWithNewOlapConnector } from "./code-utils";

describe("updateRillYAMLBlobWithNewOlapConnector", () => {
  it("should add a new `olap_connector` key to a blank file", () => {
    const updatedBlob = updateRillYAMLBlobWithNewOlapConnector(
      "",
      "clickhouse",
    );
    expect(updatedBlob).toBe("olap_connector: clickhouse\n");
  });

  it("should add a new `olap_connector` key to a file with other keys", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n`;
    const updatedBlob = updateRillYAMLBlobWithNewOlapConnector(
      existingBlob,
      "clickhouse",
    );
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nolap_connector: clickhouse\n`,
    );
  });

  it("should update the `olap_connector` key in a file with an existing `olap_connector` key", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n\nolap_connector: snowflake\n`;
    const updatedBlob = updateRillYAMLBlobWithNewOlapConnector(
      existingBlob,
      "clickhouse",
    );
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nolap_connector: clickhouse\n`,
    );
  });
});
