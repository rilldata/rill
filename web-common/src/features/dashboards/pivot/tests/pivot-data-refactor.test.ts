import {
  buildFinalPivotStateDetails,
  createPivotDataCache,
  getPivotSkeletonForPage,
  syncPivotCacheToConfig,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-data-assembly";
import {
  type PivotBaseQueryPlan,
  applyOutermostRowLimit,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-query-plan";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { describe, expect, it } from "vitest";

function config(
  overrides: Partial<PivotDataStoreConfig> = {},
): PivotDataStoreConfig {
  return {
    measureNames: ["measure"],
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    allMeasures: [],
    allDimensions: [],
    whereFilter: createAndExpression([]),
    pivot: {
      activeCell: null,
      columnPage: 1,
      columns: [],
      enableComparison: false,
      expanded: {},
      rowPage: 1,
      rows: [],
      sorting: [],
      tableMode: "nest",
      showTotalsColumn: true,
      showTotalsRow: true,
    },
    time: {
      timeStart: undefined,
      timeEnd: undefined,
      timeZone: "UTC",
      timeDimension: "",
    },
    comparisonTime: undefined,
    enableComparison: false,
    searchText: undefined,
    isFlat: false,
    ...overrides,
  };
}

function plan(overrides: Partial<PivotBaseQueryPlan> = {}): PivotBaseQueryPlan {
  return {
    anchorDimension: "country",
    configKey: "key",
    displayTotalsRow: true,
    effectiveOutermostLimit: undefined,
    isMeasureSortAccessor: false,
    measureBody: [{ name: "measure" }],
    rowAxisLimitToQuery: "50",
    rowOffset: 0,
    rowPage: 1,
    sortAccessor: undefined,
    sortFilteredMeasureBody: [{ name: "measure" }],
    sortPivotBy: [],
    timeRange: {
      start: undefined,
      end: undefined,
    },
    whereFilter: createAndExpression([]),
    ...overrides,
  };
}

describe("applyOutermostRowLimit", () => {
  it("trims the limit+1 row and reports more rows", () => {
    const result = applyOutermostRowLimit(
      config({
        pivot: {
          ...config().pivot,
          rowLimit: 2,
        },
      }),
      plan({ effectiveOutermostLimit: 2 }),
      ["US", "CA", "MX"],
      [{ country: "US" }, { country: "CA" }, { country: "MX" }],
    );

    expect(result.hasMoreRows).toBe(true);
    expect(result.rowDimensionValues).toEqual(["US", "CA"]);
    expect(result.axesRowTotals).toEqual([
      { country: "US" },
      { country: "CA" },
    ]);
  });

  it("does not trim flat table row values", () => {
    const result = applyOutermostRowLimit(
      config({ isFlat: true }),
      plan({ effectiveOutermostLimit: 2 }),
      ["US", "CA", "MX"],
      [{ country: "US" }, { country: "CA" }, { country: "MX" }],
    );

    expect(result.hasMoreRows).toBe(false);
    expect(result.rowDimensionValues).toEqual(["US", "CA", "MX"]);
  });
});

describe("pivot data assembly helpers", () => {
  it("accumulates skeleton rows when a later row page is processed", () => {
    const cache = createPivotDataCache();
    cache.lastPivotData = [{ country: "US" }];
    cache.lastProcessedRowPage = 1;

    const skeleton = getPivotSkeletonForPage(
      config({
        pivot: {
          ...config().pivot,
          rowPage: 2,
        },
      }),
      cache,
      [{ country: "CA" }],
    );

    expect(skeleton).toEqual([{ country: "US" }, { country: "CA" }]);
  });

  it("resets page accumulation when the config key changes", () => {
    const cache = createPivotDataCache();
    cache.lastProcessedConfigKey = "old";
    cache.lastProcessedRowPage = 4;

    syncPivotCacheToConfig(cache, "new");

    expect(cache.lastProcessedConfigKey).toBe("new");
    expect(cache.lastProcessedRowPage).toBe(0);
  });

  it("adds an outer show-more row and keeps row data open ended", () => {
    const finalState = buildFinalPivotStateDetails({
      anchorDimension: "country",
      columnDimensionAxes: {},
      config: config({
        pivot: {
          ...config().pivot,
          rowLimit: 5,
        },
      }),
      data: [{ country: "US" }] as PivotDataRow[],
      hasMoreRows: true,
      isCellDataEmpty: false,
      rowDimensionValues: ["US"],
      rowOffset: 0,
    });

    expect(finalState.reachedEndForRowData).toBe(false);
    expect(finalState.data).toEqual([
      { country: "US" },
      {
        country: "__rill_type_SHOW_MORE_BUTTON",
        __currentLimit: 5,
      },
    ]);
  });
});
