import { createPivotBaseQueryPlan } from "@rilldata/web-common/features/dashboards/pivot/pivot-query-plan";
import type { PivotDataStoreConfig } from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { describe, expect, it } from "vitest";

function getConfig(
  overrides: Partial<PivotDataStoreConfig> = {},
): PivotDataStoreConfig {
  return {
    measureNames: ["impressions"],
    rowDimensionNames: ["publisher"],
    colDimensionNames: [],
    allMeasures: [],
    allDimensions: [],
    whereFilter: createAndExpression([]),
    pivot: {
      rows: [],
      columns: [],
      expanded: {},
      sorting: [],
      columnPage: 1,
      rowPage: 1,
      enableComparison: false,
      tableMode: "nest",
      activeCell: null,
      showTotalsColumn: true,
      showTotalsRow: true,
    },
    time: {
      timeStart: "2022-01-01T00:00:00.000Z",
      timeEnd: "2022-02-01T00:00:00.000Z",
      timeZone: "UTC",
      timeDimension: "timestamp",
    },
    enableComparison: false,
    comparisonTime: undefined,
    searchText: undefined,
    isFlat: false,
    ...overrides,
  };
}

describe("createPivotBaseQueryPlan", () => {
  it("preserves excluded filters when applying row search", () => {
    const plan = createPivotBaseQueryPlan(
      getConfig({
        whereFilter: createAndExpression([
          createInExpression("publisher", ["Google"], true),
        ]),
        searchText: "Google",
      }),
      undefined,
    );

    expect(convertExpressionToFilterParam(plan.whereFilter)).toEqual(
      "publisher LIKE '%Google%' AND publisher NIN ('Google')",
    );
  });
});
