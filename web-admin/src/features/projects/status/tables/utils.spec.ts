import { describe, expect, it } from "vitest";
import {
  filterTemporaryTables,
  parseSizeForSorting,
  compareSizesDescending,
  formatLogTime,
  getLogLevelClass,
  getLogLevelLabel,
  formatModelSize,
  isModelPartitioned,
  isModelIncremental,
  hasModelErroredPartitions,
  shouldFilterByErrored,
  shouldFilterByPending,
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

    it("returns number as-is", () => {
      expect(parseSizeForSorting(1024)).toBe(1024);
      expect(parseSizeForSorting(0)).toBe(-1); // 0 is falsy
    });

    it("parses string to number", () => {
      expect(parseSizeForSorting("1024")).toBe(1024);
      expect(parseSizeForSorting("999999")).toBe(999999);
    });
  });

  describe("compareSizesDescending", () => {
    it("sorts larger sizes first (descending)", () => {
      expect(compareSizesDescending(100, 200)).toBeGreaterThan(0);
      expect(compareSizesDescending(200, 100)).toBeLessThan(0);
      expect(compareSizesDescending(100, 100)).toBe(0);
    });

    it("handles string sizes", () => {
      expect(compareSizesDescending("100", "200")).toBeGreaterThan(0);
      expect(compareSizesDescending("200", "100")).toBeLessThan(0);
    });

    it("handles mixed string and number", () => {
      expect(compareSizesDescending(100, "200")).toBeGreaterThan(0);
      expect(compareSizesDescending("200", 100)).toBeLessThan(0);
    });

    it("puts undefined/invalid values last", () => {
      expect(compareSizesDescending(100, undefined)).toBeLessThan(0);
      expect(compareSizesDescending(undefined, 100)).toBeGreaterThan(0);
      expect(compareSizesDescending(undefined, undefined)).toBe(0);
    });
  });

  describe("formatLogTime", () => {
    it("returns empty string for undefined", () => {
      expect(formatLogTime(undefined)).toBe("");
    });

    it("returns empty string for empty string", () => {
      expect(formatLogTime("")).toBe("");
    });

    it("formats ISO timestamp to locale time", () => {
      // Use a fixed timestamp to test formatting
      const result = formatLogTime("2024-01-15T14:30:45.000Z");
      // Result will vary by locale, but should contain time components
      expect(result).toMatch(/\d{1,2}:\d{2}:\d{2}/);
    });
  });

  describe("getLogLevelClass", () => {
    const testCases: [string | undefined, string][] = [
      ["LOG_LEVEL_ERROR", "text-red-600"],
      ["LOG_LEVEL_FATAL", "text-red-600"],
      ["LOG_LEVEL_WARN", "text-yellow-600"],
      ["LOG_LEVEL_INFO", "text-fg-muted"],
      ["LOG_LEVEL_DEBUG", "text-fg-muted"],
      [undefined, "text-fg-muted"],
      ["UNKNOWN", "text-fg-muted"],
    ];

    for (const [level, expectedClass] of testCases) {
      it(`getLogLevelClass(${JSON.stringify(level)}) = ${expectedClass}`, () => {
        expect(getLogLevelClass(level)).toBe(expectedClass);
      });
    }
  });

  describe("getLogLevelLabel", () => {
    const testCases: [string | undefined, string][] = [
      ["LOG_LEVEL_ERROR", "ERROR"],
      ["LOG_LEVEL_FATAL", "FATAL"],
      ["LOG_LEVEL_WARN", "WARN"],
      ["LOG_LEVEL_INFO", "INFO"],
      ["LOG_LEVEL_DEBUG", "DEBUG"],
      [undefined, "INFO"],
      ["UNKNOWN", "INFO"],
    ];

    for (const [level, expectedLabel] of testCases) {
      it(`getLogLevelLabel(${JSON.stringify(level)}) = ${expectedLabel}`, () => {
        expect(getLogLevelLabel(level)).toBe(expectedLabel);
      });
    }
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
});
