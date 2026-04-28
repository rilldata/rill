import { describe, it, expect, beforeAll } from "vitest";
import {
  getSchemaNameFromDriver,
  getConnectorSchema,
  getBackendConnectorName,
  isMultiStepConnector,
  hasExplorerStep,
  getFormHeight,
  shouldShowSkipLink,
  toConnectorDriver,
  multiStepFormSchemas,
  populateSchemaCache,
} from "./connector-schemas";
import type { MultiStepFormSchema } from "../../templates/schemas/types";

// Test fixtures. The static schema imports were removed in PR 3 — at runtime
// schemas come from the `ListTemplates` RPC, so these specs seed the cache with
// just enough shape for the helper functions under test. Each fixture mirrors
// the real schema's category/step/auth setup; details that aren't exercised
// here (placeholders, hints, validation) are omitted.
const testSchemas: Record<string, MultiStepFormSchema> = {
  s3: {
    type: "object",
    title: "Amazon S3",
    "x-category": "objectStore",
    properties: {
      access_key: { type: "string", "x-step": "connector" },
      path: { type: "string", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  gcs: {
    type: "object",
    title: "Google Cloud Storage",
    "x-category": "objectStore",
    properties: {
      key: { type: "string", "x-step": "connector" },
      path: { type: "string", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  azure: {
    type: "object",
    title: "Azure Blob Storage",
    "x-category": "objectStore",
    properties: {
      account: { type: "string", "x-step": "connector" },
      path: { type: "string", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  postgres: {
    type: "object",
    title: "Postgres",
    "x-category": "sqlStore",
    properties: {
      host: { type: "string", "x-step": "connector" },
      sql: { type: "string", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  mysql: {
    type: "object",
    title: "MySQL",
    "x-category": "sqlStore",
    properties: {
      host: { type: "string", "x-step": "connector" },
      sql: { type: "string", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  snowflake: {
    type: "object",
    title: "Snowflake",
    "x-category": "warehouse",
    "x-form-height": "tall",
    properties: {
      account: { type: "string", "x-step": "connector" },
      sql: { type: "string", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  bigquery: {
    type: "object",
    title: "BigQuery",
    "x-category": "warehouse",
    properties: {
      project_id: { type: "string", "x-step": "connector" },
      sql: { type: "string", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  salesforce: {
    type: "object",
    title: "Salesforce",
    "x-category": "warehouse",
    properties: {
      username: { type: "string", "x-step": "connector" },
    },
  } as unknown as MultiStepFormSchema,
  sqlite: {
    type: "object",
    title: "SQLite",
    "x-category": "sqlStore",
    properties: {
      path: { type: "string", "x-step": "connector" },
    },
  } as unknown as MultiStepFormSchema,
  clickhouse: {
    type: "object",
    title: "ClickHouse",
    "x-category": "olap",
    properties: {
      host: { type: "string", "x-step": "connector" },
    },
  } as unknown as MultiStepFormSchema,
  duckdb: {
    type: "object",
    title: "DuckDB",
    "x-category": "olap",
    properties: {
      path: { type: "string", "x-step": "connector" },
    },
  } as unknown as MultiStepFormSchema,
  // x-driver override: schema name differs from the backend driver name
  motherduck: {
    type: "object",
    title: "MotherDuck",
    "x-category": "olap",
    "x-driver": "duckdb",
    properties: {
      token: { type: "string", "x-step": "connector" },
    },
  } as unknown as MultiStepFormSchema,
};

describe("connector-schemas", () => {
  beforeAll(() => {
    populateSchemaCache(testSchemas);
  });

  describe("getSchemaNameFromDriver", () => {
    it("returns driver name when it directly matches a schema name", () => {
      expect(getSchemaNameFromDriver("postgres")).toBe("postgres");
      expect(getSchemaNameFromDriver("mysql")).toBe("mysql");
      expect(getSchemaNameFromDriver("s3")).toBe("s3");
      expect(getSchemaNameFromDriver("gcs")).toBe("gcs");
    });

    it("returns schema name for drivers with x-driver override (when no direct match)", () => {
      // motherduck has x-driver: "duckdb"; reverse lookup of "duckdb" should
      // return "duckdb" because that's a direct schema name match (which
      // takes precedence over x-driver overrides).
      expect(getSchemaNameFromDriver("duckdb")).toBe("duckdb");
    });

    it("returns the driver name as fallback for unknown drivers", () => {
      expect(getSchemaNameFromDriver("unknown_driver")).toBe("unknown_driver");
      expect(getSchemaNameFromDriver("custom_connector")).toBe(
        "custom_connector",
      );
    });

    it("handles all registered schema names", () => {
      const schemaNames = Object.keys(multiStepFormSchemas);
      for (const name of schemaNames) {
        const result = getSchemaNameFromDriver(name);
        expect(result).toBe(name);
      }
    });
  });

  describe("getConnectorSchema", () => {
    it("returns schema for valid connector names", () => {
      expect(getConnectorSchema("postgres")).not.toBeNull();
      expect(getConnectorSchema("mysql")).not.toBeNull();
      expect(getConnectorSchema("s3")).not.toBeNull();
    });

    it("returns null for invalid connector names", () => {
      expect(getConnectorSchema("nonexistent")).toBeNull();
      expect(getConnectorSchema("")).toBeNull();
    });

    it("returns schema with properties", () => {
      const schema = getConnectorSchema("postgres");
      expect(schema?.properties).toBeDefined();
      expect(Object.keys(schema?.properties ?? {}).length).toBeGreaterThan(0);
    });
  });

  describe("getBackendConnectorName", () => {
    it("returns schema name when no x-driver is specified", () => {
      expect(getBackendConnectorName("postgres")).toBe("postgres");
      expect(getBackendConnectorName("mysql")).toBe("mysql");
    });

    it("returns x-driver value when specified in schema", () => {
      expect(getBackendConnectorName("motherduck")).toBe("duckdb");
    });

    it("returns schema name for unknown connectors", () => {
      expect(getBackendConnectorName("unknown")).toBe("unknown");
    });
  });

  describe("isMultiStepConnector", () => {
    it("returns true for object store connectors", () => {
      expect(isMultiStepConnector(getConnectorSchema("s3"))).toBe(true);
      expect(isMultiStepConnector(getConnectorSchema("gcs"))).toBe(true);
      expect(isMultiStepConnector(getConnectorSchema("azure"))).toBe(true);
    });

    it("returns false for non-object store connectors", () => {
      expect(isMultiStepConnector(getConnectorSchema("postgres"))).toBe(false);
      expect(isMultiStepConnector(getConnectorSchema("mysql"))).toBe(false);
    });

    it("returns false for null schema", () => {
      expect(isMultiStepConnector(null)).toBe(false);
    });
  });

  describe("hasExplorerStep", () => {
    it("returns true for warehouse and SQL store connectors with explorer step", () => {
      expect(hasExplorerStep(getConnectorSchema("snowflake"))).toBe(true);
      expect(hasExplorerStep(getConnectorSchema("postgres"))).toBe(true);
    });

    it("returns false for object store connectors", () => {
      expect(hasExplorerStep(getConnectorSchema("s3"))).toBe(false);
    });

    it("returns false for null schema", () => {
      expect(hasExplorerStep(null)).toBe(false);
    });
  });

  describe("getFormHeight", () => {
    const FORM_HEIGHT_TALL = "max-h-[40rem] min-h-[40rem]";
    const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";

    it("returns tall height for schemas with x-form-height: tall", () => {
      expect(getFormHeight(getConnectorSchema("snowflake"))).toBe(
        FORM_HEIGHT_TALL,
      );
    });

    it("returns default height for schemas without x-form-height", () => {
      expect(getFormHeight(getConnectorSchema("postgres"))).toBe(
        FORM_HEIGHT_DEFAULT,
      );
    });

    it("returns default height for null schema", () => {
      expect(getFormHeight(null)).toBe(FORM_HEIGHT_DEFAULT);
    });
  });

  describe("shouldShowSkipLink", () => {
    it("returns true for non-OLAP connector on connector step", () => {
      expect(shouldShowSkipLink("connector", "postgres", null, false)).toBe(
        true,
      );
      expect(shouldShowSkipLink("connector", "s3", null, false)).toBe(true);
    });

    it("returns false for OLAP connectors", () => {
      expect(shouldShowSkipLink("connector", "clickhouse", null, true)).toBe(
        false,
      );
      expect(shouldShowSkipLink("connector", "duckdb", null, true)).toBe(false);
    });

    it("returns false when not on connector step", () => {
      expect(shouldShowSkipLink("source", "postgres", null, false)).toBe(false);
      expect(shouldShowSkipLink("explorer", "postgres", null, false)).toBe(
        false,
      );
    });

    it("returns false when connectorInstanceName is set", () => {
      expect(
        shouldShowSkipLink("connector", "postgres", "my-connector", false),
      ).toBe(false);
    });

    it("returns false for excluded connectors (salesforce, sqlite)", () => {
      expect(shouldShowSkipLink("connector", "salesforce", null, false)).toBe(
        false,
      );
      expect(shouldShowSkipLink("connector", "sqlite", null, false)).toBe(
        false,
      );
    });
  });

  describe("toConnectorDriver", () => {
    it("returns null for unknown schema names", () => {
      expect(toConnectorDriver("nonexistent")).toBeNull();
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
