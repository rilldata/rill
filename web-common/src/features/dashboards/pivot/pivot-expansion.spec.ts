import { describe, expect, it } from "vitest";
import { LOADING_CELL } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { buildExpandKey } from "./pivot-expand-keys";
import {
  addExpandedDataToPivot,
  getValuesForExpandedKey,
} from "./pivot-expansion";
import { type PivotDataRow, type PivotDataStoreConfig } from "./types";

function getConfig(
  showTotalsRow: boolean,
  rowDimensionNames = ["publisher", "campaign"],
): PivotDataStoreConfig {
  return {
    measureNames: ["impressions"],
    rowDimensionNames,
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

describe("pivot expansion (value-based keys)", () => {
  it("fills the node matched by dimension value, regardless of totals row", () => {
    for (const showTotals of [false, true]) {
      const tableData: PivotDataRow[] = [
        { publisher: "A", subRows: [{ publisher: LOADING_CELL }] },
        { publisher: "B", subRows: [{ publisher: LOADING_CELL }] },
      ];

      addExpandedDataToPivot(
        getConfig(showTotals),
        tableData,
        ["publisher", "campaign"],
        {},
        [
          {
            isFetching: false,
            expandIndex: buildExpandKey(["B"]),
            rowDimensionValues: ["B"],
            totals: [{ campaign: "campaign-1", impressions: 10 }],
            data: [],
          },
        ],
      );

      // Row "A" is untouched; row "B" (matched by value, not position) is filled.
      expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
      expect(tableData[1].subRows?.[0]).toMatchObject({
        publisher: "campaign-1",
        campaign: "campaign-1",
        impressions: 10,
      });
    }
  });

  it("resolves a nested value path to the correct deep node", () => {
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [
          { publisher: "camp1", subRows: [{ publisher: LOADING_CELL }] },
        ],
      },
    ];

    addExpandedDataToPivot(
      getConfig(true, ["publisher", "campaign", "adgroup"]),
      tableData,
      ["publisher", "campaign", "adgroup"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(["A", "camp1"]),
          rowDimensionValues: ["A", "camp1"],
          totals: [{ adgroup: "ad-1", impressions: 5 }],
          data: [],
        },
      ],
    );

    expect(tableData[0].subRows?.[0]?.subRows?.[0]).toMatchObject({
      publisher: "ad-1",
      adgroup: "ad-1",
      impressions: 5,
    });
  });

  it("does nothing when the value path matches no row", () => {
    const tableData: PivotDataRow[] = [
      { publisher: "A", subRows: [{ publisher: LOADING_CELL }] },
    ];
    addExpandedDataToPivot(
      getConfig(false),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(["does-not-exist"]),
          rowDimensionValues: ["does-not-exist"],
          totals: [{ campaign: "c", impressions: 1 }],
          data: [],
        },
      ],
    );
    expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
  });
});

describe("getValuesForExpandedKey", () => {
  const tableData: PivotDataRow[] = [
    {
      publisher: "A",
      subRows: [{ publisher: "camp1" }, { publisher: "camp2" }],
    },
    { publisher: "B", subRows: [{ publisher: "camp3" }] },
  ];

  it("returns the actual values along the matched path", () => {
    expect(
      getValuesForExpandedKey(
        tableData,
        ["publisher", "campaign"],
        buildExpandKey(["B"]),
      ),
    ).toEqual(["B"]);
    expect(
      getValuesForExpandedKey(
        tableData,
        ["publisher", "campaign"],
        buildExpandKey(["A", "camp2"]),
      ),
    ).toEqual(["A", "camp2"]);
  });

  it("stops at the deepest resolvable segment", () => {
    expect(
      getValuesForExpandedKey(
        tableData,
        ["publisher", "campaign"],
        buildExpandKey(["A", "missing"]),
      ),
    ).toEqual(["A"]);
  });
});
