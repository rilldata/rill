/**
 * The pivot filter builders serve two callers with opposite needs:
 *
 * - Click-to-filter reads the returned expression back to decide which
 *   dimension values to toggle on the dashboard, so the expression must
 *   describe only the clicked element. Including the global where filter here
 *   made deselecting a pivot cell also toggle off unrelated global filters.
 * - The rows viewer queries actual rows with the expression, so it must honour
 *   the global where filter on top of the cell's own filters.
 *
 * These tests pin both halves of that contract.
 */
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { buildFinalPivotStateDetails } from "./pivot-data-assembly";
import {
  getFiltersForColumnHeader,
  getFiltersForRowData,
  getFiltersForRowHeader,
} from "./pivot-row-selection";
import { getFiltersForCell, getFiltersFromRow } from "./pivot-utils";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

/** An unrelated global filter, as if the user had picked it from the filter bar. */
const GLOBAL_FILTER = createAndExpression([
  createInExpression("domain", ["google.com"]),
]);

function makeConfig(
  overrides: Partial<PivotDataStoreConfig> = {},
): PivotDataStoreConfig {
  return {
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    measureNames: ["revenue"],
    isFlat: true,
    whereFilter: GLOBAL_FILTER,
    time: { timeDimension: "", timeStart: undefined, timeEnd: undefined },
    pivot: {
      activeCell: null,
      rowPage: 1,
      showTotalsRow: false,
      sorting: [],
    },
    ...overrides,
  } as unknown as PivotDataStoreConfig;
}

/** Dimension names referenced by the top-level AND expression, in order. */
function identsIn(expr: V1Expression | undefined): string[] {
  return (expr?.cond?.exprs ?? [])
    .map((e) => e.cond?.exprs?.[0]?.ident)
    .filter((ident): ident is string => ident !== undefined);
}

const FLAT_DATA: PivotDataRow[] = [
  { country: "US", revenue: 100 },
  { country: "UK", revenue: 200 },
];

describe("pivot filter builders exclude the global where filter", () => {
  it("getFiltersForRowData", () => {
    const result = getFiltersForRowData(makeConfig(), FLAT_DATA[0]);

    expect(identsIn(result.filters)).toEqual(["country"]);
  });

  it("getFiltersFromRow", () => {
    const result = getFiltersFromRow(makeConfig(), FLAT_DATA[0], "revenue");

    expect(identsIn(result.filters)).toEqual(["country"]);
  });

  it("getFiltersForRowHeader", () => {
    const result = getFiltersForRowHeader(makeConfig(), "0", FLAT_DATA);

    expect(identsIn(result.filters)).toEqual(["country"]);
  });

  it("getFiltersForColumnHeader", () => {
    const config = makeConfig({ colDimensionNames: ["region"] });

    const result = getFiltersForColumnHeader(config, { region: "NA" });

    expect(identsIn(result.filters)).toEqual(["region"]);
  });

  it("getFiltersForCell on a flat table", () => {
    const result = getFiltersForCell(
      makeConfig(),
      "0",
      "revenue",
      {},
      FLAT_DATA,
    );

    expect(identsIn(result.filters)).toEqual(["country"]);
  });

  it("getFiltersForCell on a nested table with column dimensions", () => {
    const config = makeConfig({
      isFlat: false,
      rowDimensionNames: ["country"],
      colDimensionNames: ["region"],
    });

    const result = getFiltersForCell(
      config,
      "0",
      "c0v0m0",
      { region: ["NA", "EU"] },
      FLAT_DATA,
    );

    expect(identsIn(result.filters)).toEqual(["country", "region"]);
  });

  it("does not merge a global filter on a dimension the pivot also filters", () => {
    // The global filter already narrows `country`; the cell's own filter must
    // not be intersected with it, otherwise click-to-filter would toggle
    // values the user selected elsewhere.
    const config = makeConfig({
      whereFilter: createAndExpression([
        createInExpression("country", ["US", "UK"]),
      ]),
    });

    const result = getFiltersFromRow(config, FLAT_DATA[0], "revenue");

    expect(identsIn(result.filters)).toEqual(["country"]);
    const countryExpr = result.filters?.cond?.exprs?.[0];
    expect(countryExpr?.cond?.exprs?.slice(1).map((e) => e.val)).toEqual([
      "US",
    ]);
  });
});

describe("rows viewer cell filters include the global where filter", () => {
  it("merges config.whereFilter into activeCellFilters", () => {
    const config = makeConfig({
      pivot: {
        activeCell: { rowId: "0", columnId: "revenue" },
        rowPage: 1,
        showTotalsRow: false,
        sorting: [],
      } as unknown as PivotDataStoreConfig["pivot"],
    });

    const { activeCellFilters } = buildFinalPivotStateDetails({
      anchorDimension: "country",
      columnDimensionAxes: {},
      config,
      data: FLAT_DATA,
      hasMoreRows: false,
      isCellDataEmpty: false,
      rowDimensionValues: ["US", "UK"],
      rowOffset: 0,
    });

    expect(identsIn(activeCellFilters?.filters).sort()).toEqual([
      "country",
      "domain",
    ]);
  });

  it("is undefined when no cell is active", () => {
    const { activeCellFilters } = buildFinalPivotStateDetails({
      anchorDimension: "country",
      columnDimensionAxes: {},
      config: makeConfig(),
      data: FLAT_DATA,
      hasMoreRows: false,
      isCellDataEmpty: false,
      rowDimensionValues: ["US", "UK"],
      rowOffset: 0,
    });

    expect(activeCellFilters).toBeUndefined();
  });
});
