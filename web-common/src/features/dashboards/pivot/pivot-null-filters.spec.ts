import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { PivotDataStoreConfig } from "@rilldata/web-common/features/dashboards/pivot/types";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  extractDimensionFiltersFromExpression,
  getFiltersForRowData,
} from "./pivot-row-selection";
import { getFiltersFromRow } from "./pivot-utils";

/** Minimal config for testing row-dimension filter generation. */
function makeConfig(
  rowDimensionNames: string[],
  isFlat = true,
): PivotDataStoreConfig {
  return {
    rowDimensionNames,
    colDimensionNames: [],
    measureNames: ["revenue"],
    isFlat,
    time: {
      timeDimension: "",
      timeStart: undefined,
      timeEnd: undefined,
    },
    whereFilter: undefined,
  } as unknown as PivotDataStoreConfig;
}

describe("getFiltersFromRow: null dimension values", () => {
  it("should produce a filter for a null dimension value", () => {
    const config = makeConfig(["country"]);
    const rowData = { country: null, revenue: 100 };

    const result = getFiltersFromRow(config, rowData, "country");

    // The filter should exist and contain an IN expression for null
    expect(result.filters).toBeDefined();
    expect(result.filters?.cond?.op).toBe(V1Operation.OPERATION_AND);

    const exprs = result.filters?.cond?.exprs ?? [];
    const countryExpr = exprs.find(
      (e) => e.cond?.exprs?.[0]?.ident === "country",
    );
    expect(countryExpr).toBeDefined();
    expect(countryExpr?.cond?.op).toBe(V1Operation.OPERATION_IN);
    expect(countryExpr?.cond?.exprs?.[1]?.val).toBeNull();
  });

  it("should produce filters for a mix of null and non-null dimensions", () => {
    const config = makeConfig(["country", "city"]);
    const rowData = { country: "US", city: null, revenue: 100 };

    const result = getFiltersFromRow(config, rowData, "city");

    expect(result.filters).toBeDefined();
    const exprs = result.filters?.cond?.exprs ?? [];

    const countryExpr = exprs.find(
      (e) => e.cond?.exprs?.[0]?.ident === "country",
    );
    const cityExpr = exprs.find((e) => e.cond?.exprs?.[0]?.ident === "city");

    expect(countryExpr).toBeDefined();
    expect(countryExpr?.cond?.exprs?.[1]?.val).toBe("US");

    expect(cityExpr).toBeDefined();
    expect(cityExpr?.cond?.exprs?.[1]?.val).toBeNull();
  });
});

describe("getFiltersForRowData: null dimension values", () => {
  it("should produce a filter for a null dimension value", () => {
    const config = makeConfig(["country"]);
    const rowData = { country: null, revenue: 100 };

    const result = getFiltersForRowData(config, rowData);

    expect(result.filters).toBeDefined();
    expect(result.filters?.cond?.op).toBe(V1Operation.OPERATION_AND);

    const exprs = result.filters?.cond?.exprs ?? [];
    const countryExpr = exprs.find(
      (e) => e.cond?.exprs?.[0]?.ident === "country",
    );
    expect(countryExpr).toBeDefined();
    expect(countryExpr?.cond?.exprs?.[1]?.val).toBeNull();
  });
});

describe("extractDimensionFiltersFromExpression: null values", () => {
  it("should include null values when extracting dimension filters", () => {
    const expr = createAndExpression([createInExpression("country", [null])]);

    const result = extractDimensionFiltersFromExpression(expr);

    expect(result).toHaveLength(1);
    expect(result[0].dimensionName).toBe("country");
    expect(result[0].values).toContain(null);
  });

  it("should include both null and string values", () => {
    const expr = createAndExpression([
      createInExpression("country", ["US", null]),
    ]);

    const result = extractDimensionFiltersFromExpression(expr);

    expect(result).toHaveLength(1);
    expect(result[0].dimensionName).toBe("country");
    expect(result[0].values).toContain("US");
    expect(result[0].values).toContain(null);
  });
});
