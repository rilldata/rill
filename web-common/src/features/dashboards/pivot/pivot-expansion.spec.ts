import { describe, expect, it } from "vitest";
import { LOADING_CELL } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { addExpandedDataToPivot } from "./pivot-expansion";
import { type PivotDataRow, type PivotDataStoreConfig } from "./types";

function getConfig(showColTotals: boolean): PivotDataStoreConfig {
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
      showRowTotals: true,
      showColTotals,
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

  it("keeps the totals row offset when totals row is visible", () => {
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
          expandIndex: "2",
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
