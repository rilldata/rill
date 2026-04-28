import { describe, it, expect } from "vitest";
import { mapResource, type DescribeEntry } from "./resource-mappers";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

/** Helper to find entries by section and label */
function find(
  entries: DescribeEntry[],
  section: string,
  label: string,
): DescribeEntry | undefined {
  return entries.find((e) => e.section === section && e.label === label);
}

/** Helper to find all entries in a section */
function findAll(entries: DescribeEntry[], section: string): DescribeEntry[] {
  return entries.filter((e) => e.section === section);
}

/** Helper to find grouped entries */
function findGrouped(
  entries: DescribeEntry[],
  section: string,
  group: string,
): DescribeEntry[] {
  return entries.filter((e) => e.section === section && e.group === group);
}

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment */
function makeResource(overrides: Record<string, any>): V1Resource {
  return {
    meta: {
      name: { kind: "rill.runtime.v1.Source", name: "test" },
      reconcileStatus: "RECONCILE_STATUS_IDLE",
      specUpdatedOn: "2025-01-01T00:00:00Z",
      stateUpdatedOn: "2025-01-01T00:00:00Z",
      ...overrides.meta,
    },
    ...overrides,
  } as V1Resource;
}
/* eslint-enable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment */

describe("resource-mappers", () => {
  describe("mapResource — basic structure", () => {
    it("returns metadata entries for any resource", () => {
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
      });
      const entries = mapResource(resource);
      const meta = findAll(entries, "Metadata");

      expect(meta.length).toBeGreaterThan(0);
      expect(find(entries, "Metadata", "Reconcile Status")).toBeDefined();
      expect(find(entries, "Metadata", "Spec Updated")).toBeDefined();
      expect(find(entries, "Metadata", "State Updated")).toBeDefined();
    });

    it("includes reconcile error when present", () => {
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
        meta: {
          name: { kind: "rill.runtime.v1.Source", name: "test" },
          reconcileStatus: "RECONCILE_STATUS_ERROR",
          reconcileError: "connection refused",
          specUpdatedOn: "2025-01-01T00:00:00Z",
          stateUpdatedOn: "2025-01-01T00:00:00Z",
        },
      });
      const entries = mapResource(resource);
      const err = find(entries, "Metadata", "Reconcile Error");
      expect(err).toBeDefined();
      expect(err!.value).toBe("connection refused");
    });

    it("includes references when present", () => {
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
        meta: {
          name: { kind: "rill.runtime.v1.Source", name: "test" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "2025-01-01T00:00:00Z",
          stateUpdatedOn: "2025-01-01T00:00:00Z",
          refs: [
            { kind: "rill.runtime.v1.Connector", name: "s3_conn" },
            { kind: "rill.runtime.v1.Model", name: "my_model" },
          ],
        },
      });
      const entries = mapResource(resource);
      const refs = find(entries, "Metadata", "References");
      expect(refs).toBeDefined();
      expect(refs!.value).toContain("s3_conn");
      expect(refs!.value).toContain("my_model");
      expect(refs!.mono).toBe(true);
    });

    it("returns empty array for resource with no known kind wrapper", () => {
      const resource = makeResource({});
      const entries = mapResource(resource);
      // Should still have metadata entries
      const meta = findAll(entries, "Metadata");
      expect(meta.length).toBeGreaterThan(0);
      // But no spec entries
      const spec = findAll(entries, "Spec");
      expect(spec.length).toBe(0);
    });
  });

  describe("mapResource — source resources", () => {
    it("extracts source spec fields", () => {
      const resource = makeResource({
        source: {
          spec: {
            sourceConnector: "s3",
            refreshSchedule: { cron: "0 * * * *" },
            timeoutSeconds: 300,
          },
        },
      });
      const entries = mapResource(resource);

      const connector = find(entries, "Spec", "Source Connector");
      expect(connector).toBeDefined();
      expect(connector!.value).toBe("s3");

      const timeout = find(entries, "Spec", "Timeout (seconds)");
      expect(timeout).toBeDefined();
      expect(timeout!.value).toBe("300");
    });
  });

  describe("mapResource — model resources", () => {
    it("extracts model spec fields including SQL as multiline", () => {
      const resource = makeResource({
        model: {
          spec: {
            sql: "SELECT * FROM raw_events WHERE ts > '2024-01-01'",
            inputConnector: "duckdb",
            outputConnector: "duckdb",
          },
        },
      });
      const entries = mapResource(resource);

      const sql = find(entries, "Spec", "Sql");
      expect(sql).toBeDefined();
      expect(sql!.value).toContain("SELECT * FROM raw_events");
      expect(sql!.multiline).toBe(true);
      expect(sql!.mono).toBe(true);

      expect(find(entries, "Spec", "Input Connector")).toBeDefined();
      expect(find(entries, "Spec", "Output Connector")).toBeDefined();
    });
  });

  describe("mapResource — metricsView resources", () => {
    it("extracts metrics view with measures and dimensions as grouped sections", () => {
      const resource = makeResource({
        metricsView: {
          spec: {
            displayName: "Revenue Metrics",
            table: "analytics",
            timeDimension: "ts",
            measures: [
              { name: "total_revenue", expression: "SUM(revenue)" },
              { name: "count", expression: "COUNT(*)" },
            ],
            dimensions: [
              { name: "country", column: "country_code" },
              { name: "device", column: "device_type" },
            ],
          },
        },
      });
      const entries = mapResource(resource);

      // Display name
      expect(find(entries, "Spec", "Display Name")?.value).toBe(
        "Revenue Metrics",
      );

      // Time dimension
      expect(find(entries, "Spec", "Time Dimension")?.value).toBe("ts");

      // Measures promoted to their own section with groups
      const measure1 = findGrouped(entries, "Measures", "total_revenue");
      expect(measure1.length).toBeGreaterThan(0);
      const exprEntry = measure1.find((e) => e.label === "Expression");
      expect(exprEntry?.value).toBe("SUM(revenue)");

      const measure2 = findGrouped(entries, "Measures", "count");
      expect(measure2.length).toBeGreaterThan(0);

      // Dimensions promoted similarly
      const dim1 = findGrouped(entries, "Dimensions", "country");
      expect(dim1.length).toBeGreaterThan(0);
      const colEntry = dim1.find((e) => e.label === "Column");
      expect(colEntry?.value).toBe("country_code");
    });

    it("handles firstDayOfWeek formatting", () => {
      const resource = makeResource({
        metricsView: {
          spec: {
            table: "t",
            timeDimension: "ts",
            firstDayOfWeek: 1,
          },
        },
      });
      const entries = mapResource(resource);
      const dow = find(entries, "Spec", "First Day of Week");
      expect(dow).toBeDefined();
      expect(dow!.value).toBe("Monday");
    });

    it("formats firstDayOfWeek 0 as Sunday", () => {
      const resource = makeResource({
        metricsView: {
          spec: { table: "t", timeDimension: "ts", firstDayOfWeek: 0 },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "First Day of Week")?.value).toBe("Sunday");
    });

    it("formats firstMonthOfYear with 1-indexed months", () => {
      const resource = makeResource({
        metricsView: {
          spec: { table: "t", timeDimension: "ts", firstMonthOfYear: 1 },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "First Month of Year")?.value).toBe(
        "January",
      );
    });

    it("formats firstMonthOfYear 4 as April", () => {
      const resource = makeResource({
        metricsView: {
          spec: { table: "t", timeDimension: "ts", firstMonthOfYear: 4 },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "First Month of Year")?.value).toBe("April");
    });

    it("formats firstMonthOfYear 12 as December", () => {
      const resource = makeResource({
        metricsView: {
          spec: { table: "t", timeDimension: "ts", firstMonthOfYear: 12 },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "First Month of Year")?.value).toBe(
        "December",
      );
    });
  });

  describe("mapResource — SKIP_KEYS filtering", () => {
    it("skips internal/noisy fields", () => {
      const resource = makeResource({
        source: {
          spec: {
            sourceConnector: "s3",
            trigger: true,
            triggerFull: true,
            specHash: "abc123",
            refsHash: "def456",
            version: 5,
          },
        },
      });
      const entries = mapResource(resource);
      const specEntries = findAll(entries, "Spec");

      expect(specEntries.find((e) => e.label === "Trigger")).toBeUndefined();
      expect(
        specEntries.find((e) => e.label === "Trigger Full"),
      ).toBeUndefined();
      expect(specEntries.find((e) => e.label === "Spec Hash")).toBeUndefined();
      expect(specEntries.find((e) => e.label === "Refs Hash")).toBeUndefined();
      expect(specEntries.find((e) => e.label === "Version")).toBeUndefined();

      // sourceConnector should still be present
      expect(find(entries, "Spec", "Source Connector")).toBeDefined();
    });
  });

  describe("mapResource — boolean formatting", () => {
    it("formats booleans as Yes/No", () => {
      const resource = makeResource({
        source: {
          spec: {
            stageChanges: true,
            streamIngestion: false,
          },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "Stage Changes")?.value).toBe("Yes");
      expect(find(entries, "Spec", "Stream Ingestion")?.value).toBe("No");
    });
  });

  describe("mapResource — empty/null value filtering", () => {
    it("skips undefined, null, and empty string values", () => {
      const resource = makeResource({
        source: {
          spec: {
            sourceConnector: "s3",
            sinkConnector: "",
            refreshSchedule: null,
          },
        },
      });
      const entries = mapResource(resource);

      expect(find(entries, "Spec", "Source Connector")).toBeDefined();
      expect(find(entries, "Spec", "Sink Connector")).toBeUndefined();
      expect(find(entries, "Spec", "Refresh Schedule")).toBeUndefined();
    });
  });

  describe("mapResource — array handling", () => {
    it("joins primitive arrays as comma-separated values", () => {
      const resource = makeResource({
        metricsView: {
          spec: {
            table: "t",
            timeDimension: "ts",
            securityRules: ["rule1", "rule2", "rule3"],
          },
        },
      });
      const entries = mapResource(resource);
      const rules = find(entries, "Spec", "Security Rules");
      expect(rules).toBeDefined();
      expect(rules!.value).toBe("rule1, rule2, rule3");
      expect(rules!.mono).toBe(true);
    });

    it("skips empty arrays", () => {
      const resource = makeResource({
        metricsView: {
          spec: {
            table: "t",
            timeDimension: "ts",
            measures: [],
            dimensions: [],
          },
        },
      });
      const entries = mapResource(resource);
      expect(findAll(entries, "Measures").length).toBe(0);
      expect(findAll(entries, "Dimensions").length).toBe(0);
    });
  });

  describe("mapResource — proto enum cleaning", () => {
    it("cleans proto enum prefixes in values", () => {
      const resource = makeResource({
        metricsView: {
          spec: {
            table: "t",
            timeDimension: "ts",
            smallestTimeGrain: "TIME_GRAIN_DAY",
          },
        },
      });
      const entries = mapResource(resource);
      const grain = find(entries, "Spec", "Smallest Time Grain");
      expect(grain).toBeDefined();
      expect(grain!.value).toBe("Day");
    });

    it("falls through to raw value when cleanEnum returns empty for UNSPECIFIED", () => {
      const resource = makeResource({
        metricsView: {
          spec: {
            table: "t",
            timeDimension: "ts",
            changeMode: "CHANGE_MODE_UNSPECIFIED",
          },
        },
      });
      const entries = mapResource(resource);
      // formatValue returns null when cleanEnum yields empty → raw value is used as fallback
      const cm = find(entries, "Spec", "Change Mode");
      expect(cm).toBeDefined();
      expect(cm!.value).toBe("CHANGE_MODE_UNSPECIFIED");
    });

    it("cleans reconcile status in metadata", () => {
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
        meta: {
          name: { kind: "rill.runtime.v1.Source", name: "test" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "2025-01-01T00:00:00Z",
          stateUpdatedOn: "2025-01-01T00:00:00Z",
        },
      });
      const entries = mapResource(resource);
      const status = find(entries, "Metadata", "Reconcile Status");
      expect(status).toBeDefined();
      expect(status!.value).toBe("Idle");
    });
  });

  describe("mapResource — label formatting", () => {
    it("uses LABEL_MAP for known keys", () => {
      const resource = makeResource({
        source: {
          spec: {
            sourceConnector: "gcs",
            refreshSchedule: { cron: "daily" },
          },
        },
      });
      const entries = mapResource(resource);
      // sourceConnector should map to "Source Connector"
      expect(find(entries, "Spec", "Source Connector")).toBeDefined();
    });

    it("converts camelCase to spaced words for unknown keys", () => {
      const resource = makeResource({
        source: {
          spec: {
            myCustomField: "value",
          },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "My Custom Field")).toBeDefined();
    });
  });

  describe("mapResource — mono formatting", () => {
    it("marks SQL-like fields as mono", () => {
      const resource = makeResource({
        model: {
          spec: {
            sql: "SELECT 1",
            inputConnector: "duckdb",
            table: "events",
          },
        },
      });
      const entries = mapResource(resource);

      expect(find(entries, "Spec", "Sql")?.mono).toBe(true);
      expect(find(entries, "Spec", "Input Connector")?.mono).toBe(true);
      expect(find(entries, "Spec", "Table")?.mono).toBe(true);
    });
  });

  describe("mapResource — date formatting", () => {
    it("formats date fields ending in On", () => {
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
        meta: {
          name: { kind: "rill.runtime.v1.Source", name: "test" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "2025-06-15T12:30:00Z",
          stateUpdatedOn: "2025-06-15T12:30:00Z",
        },
      });
      const entries = mapResource(resource);
      const specUpdated = find(entries, "Metadata", "Spec Updated");
      expect(specUpdated).toBeDefined();
      // Should be a formatted date string, not raw ISO
      expect(specUpdated!.value).not.toBe("2025-06-15T12:30:00Z");
      expect(specUpdated!.value.length).toBeGreaterThan(0);
    });

    it("handles empty date gracefully", () => {
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
        meta: {
          name: { kind: "rill.runtime.v1.Source", name: "test" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "",
          stateUpdatedOn: "",
        },
      });
      const entries = mapResource(resource);
      // Empty dates should still create entries but with empty values
      // (the push function filters out empty strings)
      const specUpdated = find(entries, "Metadata", "Spec Updated");
      expect(specUpdated).toBeUndefined();
    });
  });

  describe("mapResource — canvas resources", () => {
    it("extracts components from canvas rows", () => {
      const componentResource: V1Resource = {
        meta: {
          name: { kind: "rill.runtime.v1.Component", name: "revenue_chart" },
        },
        component: {
          spec: {
            renderer: "vega_lite",
            rendererProperties: {
              metricsView: "revenue_mv",
              chartType: "bar",
            },
          },
        },
      } as unknown as V1Resource;

      const resource = makeResource({
        canvas: {
          spec: {
            displayName: "My Dashboard",
            rows: [
              {
                height: 400,
                heightUnit: "px",
                items: [{ component: "revenue_chart", width: 12 }],
              },
            ],
          },
        },
        meta: {
          name: { kind: "rill.runtime.v1.Canvas", name: "dashboard" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "2025-01-01T00:00:00Z",
          stateUpdatedOn: "2025-01-01T00:00:00Z",
        },
      });

      const entries = mapResource(resource, [componentResource]);

      // Components section should exist
      const compEntries = findAll(entries, "Components");
      expect(compEntries.length).toBeGreaterThan(0);

      // Should have renderer info from the component resource
      const chartGroup = findGrouped(entries, "Components", "revenue_chart");
      expect(chartGroup.length).toBeGreaterThan(0);

      const renderer = chartGroup.find((e) => e.label === "Renderer");
      expect(renderer?.value).toBe("vega_lite");

      const mv = chartGroup.find((e) => e.label === "Metrics View");
      expect(mv?.value).toBe("revenue_mv");

      // Row and height info
      const row = chartGroup.find((e) => e.label === "Row");
      expect(row?.value).toBe("1");

      const rowHeight = chartGroup.find((e) => e.label === "Row Height");
      expect(rowHeight?.value).toBe("400px");
    });

    it("handles canvas with no rows", () => {
      const resource = makeResource({
        canvas: {
          spec: {
            displayName: "Empty Dashboard",
            rows: [],
          },
        },
        meta: {
          name: { kind: "rill.runtime.v1.Canvas", name: "empty" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "2025-01-01T00:00:00Z",
          stateUpdatedOn: "2025-01-01T00:00:00Z",
        },
      });
      const entries = mapResource(resource);
      const compEntries = findAll(entries, "Components");
      expect(compEntries.length).toBe(0);
    });

    it("uses snake_case metrics_view from renderer properties", () => {
      const componentResource: V1Resource = {
        meta: {
          name: { kind: "rill.runtime.v1.Component", name: "chart1" },
        },
        component: {
          spec: {
            renderer: "vega_lite",
            rendererProperties: {
              metrics_view: "snake_case_mv",
            },
          },
        },
      } as unknown as V1Resource;

      const resource = makeResource({
        canvas: {
          spec: {
            rows: [{ items: [{ component: "chart1" }] }],
          },
        },
        meta: {
          name: { kind: "rill.runtime.v1.Canvas", name: "test" },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
          specUpdatedOn: "2025-01-01T00:00:00Z",
          stateUpdatedOn: "2025-01-01T00:00:00Z",
        },
      });

      const entries = mapResource(resource, [componentResource]);
      const group = findGrouped(entries, "Components", "chart1");
      const mv = group.find((e) => e.label === "Metrics View");
      expect(mv?.value).toBe("snake_case_mv");
    });
  });

  describe("mapResource — nested objects", () => {
    it("flattens nested objects with dot-separated labels", () => {
      const resource = makeResource({
        source: {
          spec: {
            inputProperties: {
              path: "s3://bucket/data",
              format: "parquet",
            },
          },
        },
      });
      const entries = mapResource(resource);
      const pathEntry = find(entries, "Spec", "Input Properties.Path");
      expect(pathEntry).toBeDefined();
      expect(pathEntry!.value).toBe("s3://bucket/data");

      const formatEntry = find(entries, "Spec", "Input Properties.Format");
      expect(formatEntry).toBeDefined();
    });
  });

  describe("mapResource — watermark expression as multiline", () => {
    it("renders watermarkExpression as multiline code", () => {
      const resource = makeResource({
        model: {
          spec: {
            watermarkExpression: "MAX(updated_at)",
          },
        },
      });
      const entries = mapResource(resource);
      const wm = find(entries, "Spec", "Watermark Expression");
      expect(wm).toBeDefined();
      expect(wm!.multiline).toBe(true);
      expect(wm!.mono).toBe(true);
      expect(wm!.value).toBe("MAX(updated_at)");
    });
  });

  describe("mapResource — kind detection", () => {
    it("processes explore resources", () => {
      const resource = makeResource({
        explore: {
          spec: {
            displayName: "Explore View",
            metricsView: "revenue",
          },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "Display Name")?.value).toBe("Explore View");
      expect(find(entries, "Spec", "Metrics View")?.value).toBe("revenue");
    });

    it("processes component resources", () => {
      const resource = makeResource({
        component: {
          spec: {
            renderer: "vega_lite",
            definedInCanvas: "my_canvas",
          },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "Renderer")).toBeDefined();
      expect(find(entries, "Spec", "Defined in Canvas")?.value).toBe(
        "my_canvas",
      );
    });

    it("processes connector resources", () => {
      const resource = makeResource({
        connector: {
          spec: {
            driver: "postgres",
            databaseSchema: "public",
          },
        },
      });
      const entries = mapResource(resource);
      expect(find(entries, "Spec", "Driver")).toBeDefined();
      expect(find(entries, "Spec", "Database Schema")?.value).toBe("public");
    });

    it("only processes the first matching kind", () => {
      // If a resource somehow has both source and model, only the first should be processed
      const resource = makeResource({
        source: { spec: { sourceConnector: "s3" } },
        model: { spec: { sql: "SELECT 1" } },
      });
      const entries = mapResource(resource);
      // source comes before model in kindKeys, so sourceConnector should exist
      expect(find(entries, "Spec", "Source Connector")).toBeDefined();
      // model sql should NOT be present (break after first match)
      expect(find(entries, "Spec", "Sql")).toBeUndefined();
    });
  });

  describe("mapResource — fallback to wrapper when no spec", () => {
    it("uses the wrapper directly when no spec property exists", () => {
      const resource = makeResource({
        theme: {
          primaryColor: "#ff0000",
          secondaryColor: "#00ff00",
        },
      });
      const entries = mapResource(resource);
      const specEntries = findAll(entries, "Spec");
      expect(specEntries.length).toBeGreaterThan(0);
    });
  });
});
