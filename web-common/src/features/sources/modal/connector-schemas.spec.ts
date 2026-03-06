import { describe, it, expect, beforeAll } from "vitest";
import {
  getSchemaNameFromDriver,
  getConnectorSchema,
  getBackendConnectorName,
  isMultiStepConnector,
  hasExplorerStep,
  getFormHeight,
  shouldShowSkipLink,
  populateSchemaCache,
} from "./connector-schemas";
import type { MultiStepFormSchema } from "../../templates/schemas/types";

// Minimal test schemas that exercise the functions without requiring the API
const testSchemas: Record<string, MultiStepFormSchema> = {
  postgres: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "Postgres",
    "x-category": "sqlStore",
    properties: {
      host: { type: "string", title: "Host", "x-step": "connector" },
      sql: { type: "string", title: "SQL", "x-step": "explorer" },
      name: { type: "string", title: "Name", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  mysql: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "MySQL",
    "x-category": "sqlStore",
    properties: {
      host: { type: "string", title: "Host", "x-step": "connector" },
      sql: { type: "string", title: "SQL", "x-step": "explorer" },
      name: { type: "string", title: "Name", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  s3: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "S3",
    "x-category": "objectStore",
    properties: {
      aws_access_key_id: {
        type: "string",
        title: "Access Key",
        "x-step": "connector",
      },
      path: { type: "string", title: "Path", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  gcs: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "GCS",
    "x-category": "objectStore",
    properties: {
      google_application_credentials: {
        type: "string",
        title: "Credentials",
        "x-step": "connector",
      },
      path: { type: "string", title: "Path", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  azure: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "Azure",
    "x-category": "objectStore",
    properties: {
      azure_storage_account: {
        type: "string",
        title: "Account",
        "x-step": "connector",
      },
      path: { type: "string", title: "Path", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  snowflake: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "Snowflake",
    "x-category": "warehouse",
    "x-form-height": "tall",
    properties: {
      account: { type: "string", title: "Account", "x-step": "connector" },
      sql: { type: "string", title: "SQL", "x-step": "explorer" },
      name: { type: "string", title: "Name", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
  clickhouse: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "ClickHouse",
    "x-category": "olap",
    properties: {
      host: { type: "string", title: "Host" },
    },
  } as unknown as MultiStepFormSchema,
  duckdb: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "DuckDB",
    "x-category": "olap",
    properties: {
      path: { type: "string", title: "Path" },
    },
  } as unknown as MultiStepFormSchema,
  salesforce: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "Salesforce",
    "x-category": "sqlStore",
    properties: {
      client_id: { type: "string", title: "Client ID", "x-step": "connector" },
    },
  } as unknown as MultiStepFormSchema,
  sqlite: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    title: "SQLite",
    "x-category": "sourceOnly",
    properties: {
      db: { type: "string", title: "Database" },
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
      for (const [schemaName, schema] of Object.entries(testSchemas)) {
        const xDriver = (schema as Record<string, unknown>)?.["x-driver"] as
          | string
          | undefined;
        if (xDriver && xDriver !== schemaName && !(xDriver in testSchemas)) {
          expect(getSchemaNameFromDriver(xDriver)).toBe(schemaName);
        }
      }
    });

    it("returns the driver name as fallback for unknown drivers", () => {
      expect(getSchemaNameFromDriver("unknown_driver")).toBe("unknown_driver");
      expect(getSchemaNameFromDriver("custom_connector")).toBe(
        "custom_connector",
      );
    });

    it("handles all registered schema names", () => {
      const schemaNames = Object.keys(testSchemas);
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
      for (const [schemaName, schema] of Object.entries(testSchemas)) {
        const expected =
          (schema as Record<string, unknown>)?.["x-driver"] ?? schemaName;
        expect(getBackendConnectorName(schemaName)).toBe(expected);
      }
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
    it("returns true for SQL store and warehouse connectors", () => {
      const snowflake = getConnectorSchema("snowflake");
      const postgres = getConnectorSchema("postgres");

      if (snowflake?.["x-category"] === "warehouse") {
        expect(hasExplorerStep(snowflake)).toBe(true);
      }
      if (postgres?.["x-category"] === "sqlStore") {
        expect(hasExplorerStep(postgres)).toBe(true);
      }
    });

    it("returns false for object store connectors", () => {
      expect(hasExplorerStep(getConnectorSchema("s3"))).toBe(false);
    });

    it("returns false for null schema", () => {
      expect(hasExplorerStep(null)).toBe(false);
    });
  });

  describe("getFormHeight", () => {
    it("returns tall height for schemas with x-form-height: tall", () => {
      const FORM_HEIGHT_TALL = "max-h-[40rem] min-h-[40rem]";

      for (const [, schema] of Object.entries(testSchemas)) {
        if ((schema as Record<string, unknown>)?.["x-form-height"] === "tall") {
          expect(getFormHeight(schema)).toBe(FORM_HEIGHT_TALL);
        }
      }
    });

    it("returns default height for schemas without x-form-height", () => {
      const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";

      for (const [, schema] of Object.entries(testSchemas)) {
        if (!(schema as Record<string, unknown>)?.["x-form-height"]) {
          expect(getFormHeight(schema)).toBe(FORM_HEIGHT_DEFAULT);
        }
      }
    });

    it("returns default height for null schema", () => {
      const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";
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
});
