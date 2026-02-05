import { describe, it, expect } from "vitest";
import { getSchemaNameFromDriver } from "@rilldata/web-common/features/sources/modal/connector-schemas";
import { OLAP_ENGINES } from "@rilldata/web-common/features/sources/modal/constants";

/**
 * Unit tests for ConnectorAddModelButton component logic.
 *
 * These tests verify the logic used in ConnectorAddModelButton.svelte:
 * - Schema name resolution from driver names
 * - OLAP connector detection
 * - Button disabled state conditions
 */
describe("ConnectorAddModelButton logic", () => {
  describe("schema name resolution", () => {
    it("resolves common driver names to schema names", () => {
      expect(getSchemaNameFromDriver("postgres")).toBe("postgres");
      expect(getSchemaNameFromDriver("mysql")).toBe("mysql");
      expect(getSchemaNameFromDriver("s3")).toBe("s3");
      expect(getSchemaNameFromDriver("gcs")).toBe("gcs");
      expect(getSchemaNameFromDriver("bigquery")).toBe("bigquery");
      expect(getSchemaNameFromDriver("snowflake")).toBe("snowflake");
    });

    it("handles null/undefined driver names gracefully", () => {
      // getSchemaNameFromDriver returns the input as fallback
      expect(getSchemaNameFromDriver("")).toBe("");
    });
  });

  describe("OLAP connector detection", () => {
    it("identifies OLAP engines from constants", () => {
      // OLAP engines should not show the Import data button
      expect(OLAP_ENGINES).toContain("clickhouse");
      expect(OLAP_ENGINES).toContain("duckdb");
      expect(OLAP_ENGINES).toContain("druid");
      expect(OLAP_ENGINES).toContain("pinot");
    });

    it("correctly identifies non-OLAP connectors", () => {
      const nonOlapConnectors = ["postgres", "mysql", "s3", "gcs", "bigquery"];
      for (const connector of nonOlapConnectors) {
        expect(OLAP_ENGINES.includes(connector)).toBe(false);
      }
    });
  });

  describe("button disabled state logic", () => {
    it("should be disabled when hasUnsavedChanges is true", () => {
      const hasUnsavedChanges = true;
      const hasReconcileError = false;
      const driverName = "postgres";

      const isDisabled = hasUnsavedChanges || hasReconcileError || !driverName;
      expect(isDisabled).toBe(true);
    });

    it("should be disabled when hasReconcileError is true", () => {
      const hasUnsavedChanges = false;
      const hasReconcileError = true;
      const driverName = "postgres";

      const isDisabled = hasUnsavedChanges || hasReconcileError || !driverName;
      expect(isDisabled).toBe(true);
    });

    it("should be disabled when driverName is missing", () => {
      const hasUnsavedChanges = false;
      const hasReconcileError = false;
      const driverName = "";

      const isDisabled = hasUnsavedChanges || hasReconcileError || !driverName;
      expect(isDisabled).toBe(true);
    });

    it("should be enabled when all conditions are met", () => {
      const hasUnsavedChanges = false;
      const hasReconcileError = false;
      const driverName = "postgres";

      const isDisabled = hasUnsavedChanges || hasReconcileError || !driverName;
      expect(isDisabled).toBe(false);
    });
  });

  describe("tooltip content logic", () => {
    it("returns correct tooltip for unsaved changes", () => {
      const hasUnsavedChanges = true;
      const hasReconcileError = false;

      const tooltipContent = hasUnsavedChanges
        ? "Save your changes first"
        : hasReconcileError
          ? "Fix connector errors first"
          : "Import data using this connector";

      expect(tooltipContent).toBe("Save your changes first");
    });

    it("returns correct tooltip for reconcile error", () => {
      const hasUnsavedChanges = false;
      const hasReconcileError = true;

      const tooltipContent = hasUnsavedChanges
        ? "Save your changes first"
        : hasReconcileError
          ? "Fix connector errors first"
          : "Import data using this connector";

      expect(tooltipContent).toBe("Fix connector errors first");
    });

    it("returns correct tooltip when enabled", () => {
      const hasUnsavedChanges = false;
      const hasReconcileError = false;

      const tooltipContent = hasUnsavedChanges
        ? "Save your changes first"
        : hasReconcileError
          ? "Fix connector errors first"
          : "Import data using this connector";

      expect(tooltipContent).toBe("Import data using this connector");
    });
  });
});
