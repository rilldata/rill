import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { QueryClient } from "@tanstack/svelte-query";
import { get, readable } from "svelte/store";
import { describe, expect, it } from "vitest";
import { LOADING_CELL } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  addExpandedDataToPivot,
  createSubTableCellQuery,
} from "./pivot-expansion";
import {
  type PivotDashboardContext,
  type PivotDataRow,
  type PivotDataStoreConfig,
} from "./types";

function getConfig(showTotalsRow: boolean): PivotDataStoreConfig {
  return {
    measureNames: ["impressions"],
    rowDimensionNames: ["publisher", "campaign"],
    colDimensionNames: [],
    allMeasures: [],
    allDimensions: [],
    whereFilter: createAndExpression([]),
    pivot: {
      rows: [],
      columns: [],
      sorting: [],
      expanded: {},
      columnPage: 1,
      rowPage: 1,
      enableComparison: true,
      tableMode: "nest",
      activeCell: null,
      showTotalsColumn: true,
      showTotalsRow,
    },
    time: {
      timeStart: undefined,
      timeEnd: undefined,
      timeZone: "UTC",
      timeDimension: "timestamp",
    },
    enableComparison: false,
    comparisonTime: undefined,
    searchText: undefined,
    isFlat: false,
  } as unknown as PivotDataStoreConfig;
}

describe("pivot expansion", () => {
  it("adds expanded rows at the correct index when totals row is hidden", () => {
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [{ publisher: LOADING_CELL }],
      },
      {
        publisher: "B",
        subRows: [{ publisher: LOADING_CELL }],
      },
    ];

    addExpandedDataToPivot(
      getConfig(false),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: "1",
          rowDimensionValues: ["B"],
          totals: [{ campaign: "campaign-1", impressions: 10 }],
          data: [],
        },
      ],
    );

    expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
    expect(tableData[1].subRows?.[0]).toMatchObject({
      publisher: "campaign-1",
      campaign: "campaign-1",
      impressions: 10,
    });
  });

  it("does not offset the expand index when totals row is visible", () => {
    // The totals row is rendered outside the table's row model, so expand
    // indices always match the position in the pivot data.
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [{ publisher: LOADING_CELL }],
      },
      {
        publisher: "B",
        subRows: [{ publisher: LOADING_CELL }],
      },
    ];

    addExpandedDataToPivot(
      getConfig(true),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: "1",
          rowDimensionValues: ["B"],
          totals: [{ campaign: "campaign-1", impressions: 10 }],
          data: [],
        },
      ],
    );

    expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
    expect(tableData[1].subRows?.[0]).toMatchObject({
      publisher: "campaign-1",
      campaign: "campaign-1",
      impressions: 10,
    });
  });
});

describe("createSubTableCellQuery", () => {
  // An expanded row whose sub rows query came back empty, for example after
  // the filter or time range changed. There are no cells to fetch, and the row
  // filter would be an `IN` with no values, which the runtime rejects.
  it("skips the query for a row with no sub rows", () => {
    const requests: unknown[] = [];
    const ctx: PivotDashboardContext = {
      runtimeClient: {
        instanceId: "default",
        queryService: {
          metricsViewAggregation: (request: unknown) => {
            requests.push(request);
            return Promise.resolve({ toJson: () => ({ data: [] }) });
          },
        },
      } as unknown as RuntimeClient,
      metricsViewName: readable("metrics_view"),
      queryClient: new QueryClient({
        defaultOptions: { queries: { retry: false } },
      }),
      enabled: readable(true),
    };
    const config = { ...getConfig(false), colDimensionNames: ["country"] };

    const query = createSubTableCellQuery(
      ctx,
      config,
      "campaign",
      { country: ["US"] },
      { c0v0m0: 10, impressions: 10 },
      [],
      createAndExpression([createInExpression("publisher", ["A"])]),
      [],
    );

    expect(get(query)).toBeNull();
    expect(requests).toHaveLength(0);
  });
});
