import { describe, it, expect } from "vitest";
import {
  getSchemaNameFromDriver,
  getConnectorSchema,
  getBackendConnectorName,
  isMultiStepConnector,
  hasExplorerStep,
  getFormHeight,
  multiStepFormSchemas,
} from "./connector-schemas";

describe("connector-schemas", () => {
  describe("getSchemaNameFromDriver", () => {
    it("returns driver name when it directly matches a schema name", () => {
      expect(getSchemaNameFromDriver("postgres")).toBe("postgres");
      expect(getSchemaNameFromDriver("mysql")).toBe("mysql");
      expect(getSchemaNameFromDriver("s3")).toBe("s3");
      expect(getSchemaNameFromDriver("gcs")).toBe("gcs");
    });

    it("returns schema name for drivers with x-driver override (when no direct match)", () => {
      // Find schemas with x-driver overrides that don't directly match another schema name
      for (const [schemaName, schema] of Object.entries(multiStepFormSchemas)) {
        const xDriver = schema?.["x-driver"];
        // Only test if x-driver is set and doesn't match an existing schema name
        // (because direct schema name matches take precedence)
        if (
          xDriver &&
          xDriver !== schemaName &&
          !(xDriver in multiStepFormSchemas)
        ) {
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
      for (const [schemaName, schema] of Object.entries(multiStepFormSchemas)) {
        const expected = schema?.["x-driver"] ?? schemaName;
        expect(getBackendConnectorName(schemaName)).toBe(expected);
      }
    });

    it("returns schema name for unknown connectors", () => {
      expect(getBackendConnectorName("unknown")).toBe("unknown");
    });
  });

  describe("isMultiStepConnector", () => {
    it("returns true for object store connectors", () => {
      const s3Schema = getConnectorSchema("s3");
      const gcsSchema = getConnectorSchema("gcs");
      const azureSchema = getConnectorSchema("azure");

      expect(isMultiStepConnector(s3Schema)).toBe(true);
      expect(isMultiStepConnector(gcsSchema)).toBe(true);
      expect(isMultiStepConnector(azureSchema)).toBe(true);
    });

    it("returns false for non-object store connectors", () => {
      const postgresSchema = getConnectorSchema("postgres");
      const mysqlSchema = getConnectorSchema("mysql");

      expect(isMultiStepConnector(postgresSchema)).toBe(false);
      expect(isMultiStepConnector(mysqlSchema)).toBe(false);
    });

    it("returns false for null schema", () => {
      expect(isMultiStepConnector(null)).toBe(false);
    });
  });

  describe("hasExplorerStep", () => {
    it("returns true for SQL store and warehouse connectors", () => {
      const snowflakeSchema = getConnectorSchema("snowflake");
      const postgresSchema = getConnectorSchema("postgres");

      // Check based on category
      if (snowflakeSchema?.["x-category"] === "warehouse") {
        expect(hasExplorerStep(snowflakeSchema)).toBe(true);
      }
      if (postgresSchema?.["x-category"] === "sqlStore") {
        expect(hasExplorerStep(postgresSchema)).toBe(true);
      }
    });

    it("returns false for object store connectors", () => {
      const s3Schema = getConnectorSchema("s3");
      expect(hasExplorerStep(s3Schema)).toBe(false);
    });

    it("returns false for null schema", () => {
      expect(hasExplorerStep(null)).toBe(false);
    });
  });

  describe("getFormHeight", () => {
    it("returns tall height for schemas with x-form-height: tall", () => {
      const FORM_HEIGHT_TALL = "max-h-[38.5rem] min-h-[38.5rem]";

      for (const [, schema] of Object.entries(multiStepFormSchemas)) {
        if (schema?.["x-form-height"] === "tall") {
          expect(getFormHeight(schema)).toBe(FORM_HEIGHT_TALL);
        }
      }
    });

    it("returns default height for schemas without x-form-height", () => {
      const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";

      for (const [, schema] of Object.entries(multiStepFormSchemas)) {
        if (!schema?.["x-form-height"]) {
          expect(getFormHeight(schema)).toBe(FORM_HEIGHT_DEFAULT);
        }
      }
    });

    it("returns default height for null schema", () => {
      const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";
      expect(getFormHeight(null)).toBe(FORM_HEIGHT_DEFAULT);
    });
  });
});
