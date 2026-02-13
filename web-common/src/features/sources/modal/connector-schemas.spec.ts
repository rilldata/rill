import { describe, expect, it } from "vitest";
import {
  getBackendConnectorName,
  getConnectorSchema,
  hasExplorerStep,
  isAiConnector,
  isMultiStepConnector,
  toConnectorDriver,
} from "@rilldata/web-common/features/sources/modal/connector-schemas";

describe("connector-schemas", () => {
  describe("isAiConnector", () => {
    it("returns true for AI connector schemas", () => {
      expect(isAiConnector(getConnectorSchema("claude"))).toBe(true);
      expect(isAiConnector(getConnectorSchema("openai"))).toBe(true);
      expect(isAiConnector(getConnectorSchema("gemini"))).toBe(true);
    });

    it("returns false for non-AI connector schemas", () => {
      expect(isAiConnector(getConnectorSchema("s3"))).toBe(false);
      expect(isAiConnector(getConnectorSchema("postgres"))).toBe(false);
      expect(isAiConnector(getConnectorSchema("bigquery"))).toBe(false);
    });

    it("returns false for null or undefined", () => {
      expect(isAiConnector(null)).toBe(false);
      expect(isAiConnector(undefined as never)).toBe(false);
    });
  });

  describe("isMultiStepConnector", () => {
    it("returns true for objectStore connectors", () => {
      expect(isMultiStepConnector(getConnectorSchema("s3"))).toBe(true);
      expect(isMultiStepConnector(getConnectorSchema("gcs"))).toBe(true);
      expect(isMultiStepConnector(getConnectorSchema("azure"))).toBe(true);
    });

    it("returns false for sqlStore and warehouse connectors", () => {
      expect(isMultiStepConnector(getConnectorSchema("postgres"))).toBe(false);
      expect(isMultiStepConnector(getConnectorSchema("bigquery"))).toBe(false);
    });

    it("returns false for AI connectors", () => {
      expect(isMultiStepConnector(getConnectorSchema("claude"))).toBe(false);
    });

    it("returns false for null", () => {
      expect(isMultiStepConnector(null)).toBe(false);
    });
  });

  describe("hasExplorerStep", () => {
    it("returns true for sqlStore connectors", () => {
      expect(hasExplorerStep(getConnectorSchema("postgres"))).toBe(true);
      expect(hasExplorerStep(getConnectorSchema("mysql"))).toBe(true);
    });

    it("returns true for warehouse connectors", () => {
      expect(hasExplorerStep(getConnectorSchema("bigquery"))).toBe(true);
      expect(hasExplorerStep(getConnectorSchema("snowflake"))).toBe(true);
    });

    it("returns false for objectStore connectors", () => {
      expect(hasExplorerStep(getConnectorSchema("s3"))).toBe(false);
    });

    it("returns false for AI connectors", () => {
      expect(hasExplorerStep(getConnectorSchema("claude"))).toBe(false);
    });

    it("returns false for null", () => {
      expect(hasExplorerStep(null)).toBe(false);
    });
  });

  describe("getBackendConnectorName", () => {
    it("returns the schema name when no x-driver is set", () => {
      expect(getBackendConnectorName("postgres")).toBe("postgres");
      expect(getBackendConnectorName("s3")).toBe("s3");
      expect(getBackendConnectorName("claude")).toBe("claude");
    });

    it("returns the input name for unknown schemas", () => {
      expect(getBackendConnectorName("nonexistent")).toBe("nonexistent");
    });
  });

  describe("toConnectorDriver", () => {
    it("returns null for unknown schema names", () => {
      expect(toConnectorDriver("nonexistent")).toBeNull();
    });

    it("sets implementsAi for AI connectors", () => {
      const claude = toConnectorDriver("claude");
      expect(claude).not.toBeNull();
      expect(claude!.name).toBe("claude");
      expect(claude!.displayName).toBe("Claude");
      expect(claude!.implementsAi).toBe(true);
      expect(claude!.implementsOlap).toBe(false);
      expect(claude!.implementsWarehouse).toBe(false);
      expect(claude!.implementsObjectStore).toBe(false);
      expect(claude!.implementsSqlStore).toBe(false);
    });

    it("sets implementsWarehouse for warehouse connectors", () => {
      const bq = toConnectorDriver("bigquery");
      expect(bq).not.toBeNull();
      expect(bq!.name).toBe("bigquery");
      expect(bq!.displayName).toBe("BigQuery");
      expect(bq!.implementsWarehouse).toBe(true);
      expect(bq!.implementsAi).toBe(false);
    });

    it("sets implementsObjectStore for object store connectors", () => {
      const s3 = toConnectorDriver("s3");
      expect(s3).not.toBeNull();
      expect(s3!.implementsObjectStore).toBe(true);
      expect(s3!.implementsAi).toBe(false);
    });

    it("sets implementsOlap for OLAP connectors", () => {
      const ch = toConnectorDriver("clickhouse");
      expect(ch).not.toBeNull();
      expect(ch!.implementsOlap).toBe(true);
      expect(ch!.implementsAi).toBe(false);
    });

  });
});
