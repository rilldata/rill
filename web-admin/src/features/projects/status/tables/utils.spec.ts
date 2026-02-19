import { describe, expect, it } from "vitest";
import {
  filterTemporaryTables,
  isLikelyView,
  parseSizeForSorting,
  compareSizes,
  formatModelSize,
  isModelPartitioned,
  isModelIncremental,
  hasModelErroredPartitions,
  shouldFilterByErrored,
  shouldFilterByPending,
  splitTablesByModel,
  applyTableFilters,
} from "./utils";
import type {
  V1OlapTableInfo,
  V1Resource,
} from "@rilldata/web-common/runtime-client";

describe("tables utils", () => {
  describe("filterTemporaryTables", () => {
    it("filters out __rill_tmp_ prefixed tables", () => {
      const tables: V1OlapTableInfo[] = [
        { name: "users" },
        { name: "__rill_tmp_123" },
        { name: "orders" },
        { name: "__rill_tmp_abc" },
      ];

      const result = filterTemporaryTables(tables);

      expect(result).toEqual([{ name: "users" }, { name: "orders" }]);
    });

    it("filters out tables with empty names", () => {
      const tables: V1OlapTableInfo[] = [
        { name: "users" },
        { name: "" },
        { name: undefined },
        { name: "orders" },
      ];

      const result = filterTemporaryTables(tables);

      expect(result).toEqual([{ name: "users" }, { name: "orders" }]);
    });

    it("returns empty array for undefined input", () => {
      expect(filterTemporaryTables(undefined)).toEqual([]);
    });

    it("returns empty array for empty input", () => {
      expect(filterTemporaryTables([])).toEqual([]);
    });

    it("keeps all tables when none are temporary", () => {
      const tables: V1OlapTableInfo[] = [
        { name: "users" },
        { name: "orders" },
        { name: "products" },
      ];

      const result = filterTemporaryTables(tables);

      expect(result).toEqual(tables);
    });
  });

  describe("isLikelyView", () => {
    it("returns true when viewFlag is true", () => {
      expect(isLikelyView(true, "1024")).toBe(true);
    });

    it("returns true when physicalSizeBytes is '-1'", () => {
      expect(isLikelyView(false, "-1")).toBe(true);
    });

    it("returns true when physicalSizeBytes is 0", () => {
      expect(isLikelyView(false, 0)).toBe(true);
    });

    it("returns true when physicalSizeBytes is '0' (string)", () => {
      expect(isLikelyView(false, "0")).toBe(true);
    });

    it("returns true when physicalSizeBytes is undefined", () => {
      expect(isLikelyView(false, undefined)).toBe(true);
    });

    it("returns false for a table with valid size", () => {
      expect(isLikelyView(false, "1024")).toBe(false);
    });
  });

  describe("parseSizeForSorting", () => {
    it("returns -1 for undefined", () => {
      expect(parseSizeForSorting(undefined)).toBe(-1);
    });

    it("returns -1 for empty string", () => {
      expect(parseSizeForSorting("")).toBe(-1);
    });

    it("returns -1 for '-1' string", () => {
      expect(parseSizeForSorting("-1")).toBe(-1);
    });

    it("returns 0 for 0", () => {
      expect(parseSizeForSorting(0)).toBe(0);
    });

    it("returns number as-is", () => {
      expect(parseSizeForSorting(1024)).toBe(1024);
    });

    it("parses string to number", () => {
      expect(parseSizeForSorting("1024")).toBe(1024);
      expect(parseSizeForSorting("999999")).toBe(999999);
    });

    it("returns -1 for non-numeric strings", () => {
      expect(parseSizeForSorting("abc")).toBe(-1);
    });
  });

  describe("compareSizes", () => {
    it("sorts smaller sizes first (ascending)", () => {
      expect(compareSizes(100, 200)).toBeLessThan(0);
      expect(compareSizes(200, 100)).toBeGreaterThan(0);
      expect(compareSizes(100, 100)).toBe(0);
    });

    it("handles string sizes", () => {
      expect(compareSizes("100", "200")).toBeLessThan(0);
      expect(compareSizes("200", "100")).toBeGreaterThan(0);
    });

    it("handles mixed string and number", () => {
      expect(compareSizes(100, "200")).toBeLessThan(0);
      expect(compareSizes("200", 100)).toBeGreaterThan(0);
    });

    it("puts undefined/invalid values first (as -1)", () => {
      expect(compareSizes(100, undefined)).toBeGreaterThan(0);
      expect(compareSizes(undefined, 100)).toBeLessThan(0);
      expect(compareSizes(undefined, undefined)).toBe(0);
    });
  });

  describe("formatModelSize", () => {
    it("returns '-' for undefined", () => {
      expect(formatModelSize(undefined)).toBe("-");
    });

    it("returns '-' for null", () => {
      expect(formatModelSize(null as unknown as undefined)).toBe("-");
    });

    it("returns '-' for '-1' string", () => {
      expect(formatModelSize("-1")).toBe("-");
    });

    it("returns '-' for negative numbers", () => {
      expect(formatModelSize(-100)).toBe("-");
    });

    it("returns '-' for NaN", () => {
      expect(formatModelSize("not a number")).toBe("-");
    });

    it("formats valid byte sizes", () => {
      expect(formatModelSize(0)).toBe("0");
      expect(formatModelSize(1024)).toBe("1.0KB");
      expect(formatModelSize("1048576")).toBe("1.0MB");
    });
  });

  describe("isModelPartitioned", () => {
    it("returns false for undefined resource", () => {
      expect(isModelPartitioned(undefined)).toBe(false);
    });

    it("returns false when no partitionsResolver", () => {
      const resource: V1Resource = {
        model: { spec: {} },
      };
      expect(isModelPartitioned(resource)).toBe(false);
    });

    it("returns true when partitionsResolver exists", () => {
      const resource: V1Resource = {
        model: { spec: { partitionsResolver: "some-resolver" } },
      };
      expect(isModelPartitioned(resource)).toBe(true);
    });
  });

  describe("isModelIncremental", () => {
    it("returns false for undefined resource", () => {
      expect(isModelIncremental(undefined)).toBe(false);
    });

    it("returns false when incremental is false", () => {
      const resource: V1Resource = {
        model: { spec: { incremental: false } },
      };
      expect(isModelIncremental(resource)).toBe(false);
    });

    it("returns true when incremental is true", () => {
      const resource: V1Resource = {
        model: { spec: { incremental: true } },
      };
      expect(isModelIncremental(resource)).toBe(true);
    });
  });

  describe("hasModelErroredPartitions", () => {
    it("returns false for undefined resource", () => {
      expect(hasModelErroredPartitions(undefined)).toBe(false);
    });

    it("returns false when no partitionsModelId", () => {
      const resource: V1Resource = {
        model: { state: { partitionsHaveErrors: true } },
      };
      expect(hasModelErroredPartitions(resource)).toBe(false);
    });

    it("returns false when partitionsHaveErrors is false", () => {
      const resource: V1Resource = {
        model: {
          state: { partitionsModelId: "123", partitionsHaveErrors: false },
        },
      };
      expect(hasModelErroredPartitions(resource)).toBe(false);
    });

    it("returns true when both conditions are met", () => {
      const resource: V1Resource = {
        model: {
          state: { partitionsModelId: "123", partitionsHaveErrors: true },
        },
      };
      expect(hasModelErroredPartitions(resource)).toBe(true);
    });
  });

  describe("shouldFilterByErrored", () => {
    it("returns true for 'errors' filter", () => {
      expect(shouldFilterByErrored("errors")).toBe(true);
    });

    it("returns false for other filters", () => {
      expect(shouldFilterByErrored("all")).toBe(false);
      expect(shouldFilterByErrored("pending")).toBe(false);
    });
  });

  describe("shouldFilterByPending", () => {
    it("returns true for 'pending' filter", () => {
      expect(shouldFilterByPending("pending")).toBe(true);
    });

    it("returns false for other filters", () => {
      expect(shouldFilterByPending("all")).toBe(false);
      expect(shouldFilterByPending("errors")).toBe(false);
    });
  });

  describe("splitTablesByModel", () => {
    it("returns empty arrays for empty inputs", () => {
      const result = splitTablesByModel([], new Map());
      expect(result.modelTables).toEqual([]);
      expect(result.externalTables).toEqual([]);
    });

    it("splits tables into model-backed and external", () => {
      const tables: V1OlapTableInfo[] = [
        { name: "users" },
        { name: "orders" },
        { name: "external_data" },
      ];
      const modelResources = new Map<string, V1Resource>([
        ["users", { meta: { name: { name: "users_model" } } }],
        ["orders", { meta: { name: { name: "orders_model" } } }],
      ]);

      const result = splitTablesByModel(tables, modelResources);

      expect(result.modelTables).toEqual([
        { name: "users" },
        { name: "orders" },
      ]);
      expect(result.externalTables).toEqual([{ name: "external_data" }]);
    });

    it("matches case-insensitively", () => {
      const tables: V1OlapTableInfo[] = [{ name: "Users" }];
      const modelResources = new Map<string, V1Resource>([
        ["users", { meta: { name: { name: "users_model" } } }],
      ]);

      const result = splitTablesByModel(tables, modelResources);

      expect(result.modelTables).toEqual([{ name: "Users" }]);
      expect(result.externalTables).toEqual([]);
    });

    it("treats all tables as external when no model resources exist", () => {
      const tables: V1OlapTableInfo[] = [
        { name: "table_a" },
        { name: "table_b" },
      ];

      const result = splitTablesByModel(tables, new Map());

      expect(result.modelTables).toEqual([]);
      expect(result.externalTables).toEqual(tables);
    });
  });

  describe("applyTableFilters", () => {
    const tables: V1OlapTableInfo[] = [
      { name: "users", physicalSizeBytes: "1024" },
      { name: "orders", physicalSizeBytes: "2048" },
      { name: "analytics_view", physicalSizeBytes: "0" },
    ];
    const viewMap = new Map<string, boolean>([
      ["users", false],
      ["orders", false],
      ["analytics_view", true],
    ]);
    const modelResources = new Map<string, V1Resource>([
      ["users", { meta: { name: { name: "users_model" } } }],
    ]);

    it("returns all tables when no filters are active", () => {
      const result = applyTableFilters(
        tables,
        "",
        "all",
        viewMap,
        modelResources,
      );
      expect(result).toEqual(tables);
    });

    it("filters by OLAP table name", () => {
      const result = applyTableFilters(
        tables,
        "orders",
        "all",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([{ name: "orders", physicalSizeBytes: "2048" }]);
    });

    it("filters by model name", () => {
      const result = applyTableFilters(
        tables,
        "users_model",
        "all",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([{ name: "users", physicalSizeBytes: "1024" }]);
    });

    it("search is case-insensitive", () => {
      const result = applyTableFilters(
        tables,
        "USERS",
        "all",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([{ name: "users", physicalSizeBytes: "1024" }]);
    });

    it("filters by type: table", () => {
      const result = applyTableFilters(
        tables,
        "",
        "table",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([
        { name: "users", physicalSizeBytes: "1024" },
        { name: "orders", physicalSizeBytes: "2048" },
      ]);
    });

    it("filters by type: view", () => {
      const result = applyTableFilters(
        tables,
        "",
        "view",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([
        { name: "analytics_view", physicalSizeBytes: "0" },
      ]);
    });

    it("combines search and type filter", () => {
      const result = applyTableFilters(
        tables,
        "users",
        "table",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([{ name: "users", physicalSizeBytes: "1024" }]);
    });

    it("returns empty array when nothing matches", () => {
      const result = applyTableFilters(
        tables,
        "nonexistent",
        "all",
        viewMap,
        modelResources,
      );
      expect(result).toEqual([]);
    });
  });
});
