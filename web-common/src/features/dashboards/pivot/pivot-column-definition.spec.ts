import type { ColumnDef } from "tanstack-table-8-svelte-5";
import { describe, expect, it } from "vitest";
import { getColumnDefForPivot } from "./pivot-column-definition";
import {
  PivotChipType,
  type PivotDataRow,
  type PivotDataStoreConfig,
} from "./types";

function config(
  overrides: Partial<Omit<PivotDataStoreConfig, "pivot">> & {
    pivot?: Partial<PivotDataStoreConfig["pivot"]>;
  },
): PivotDataStoreConfig {
  const { pivot, ...rest } = overrides;
  return {
    measureNames: ["impressions"],
    rowDimensionNames: ["country"],
    colDimensionNames: ["device"],
    allMeasures: [
      {
        name: "impressions",
        displayName: "Impressions",
        formatPreset: "humanize",
      },
    ],
    allDimensions: [
      { name: "country", displayName: "Country" },
      { name: "device", displayName: "Device" },
      { name: "os", displayName: "OS" },
    ],
    whereFilter: { cond: { op: "OPERATION_AND", exprs: [] } },
    pivot: {
      columns: [],
      rows: [],
      expanded: {},
      sorting: [],
      columnPage: 1,
      rowPage: 1,
      enableComparison: false,
      tableMode: "nest",
      activeCell: null,
      showTotalsColumn: true,
      showTotalsRow: true,
      ...pivot,
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
    ...rest,
  };
}

// The header of every column with the headers of its sub-columns: one entry
// per header row, top to bottom.
type HeaderTree = [header: unknown, columns?: HeaderTree[]];

function headerTree(defs: ColumnDef<PivotDataRow>[]): HeaderTree[] {
  return defs.map((def) => {
    const columns = (def as { columns?: ColumnDef<PivotDataRow>[] }).columns;
    return columns ? [def.header, headerTree(columns)] : [def.header];
  });
}

describe("getColumnDefForPivot", () => {
  it("labels the row-totals group next to the column dimension's values", () => {
    const defs = getColumnDefForPivot(
      config({}),
      { device: ["Desktop", "Mobile"] },
      { impressions: 30, c0v0m0: 10, c0v1m0: 20 },
    );

    expect(headerTree(defs)).toStrictEqual([
      // The corner cell names the column dimension above the row dimension.
      ["Device", [["Country"]]],
      ["Total", [["Impressions"]]],
      ["Desktop", [["Impressions"]]],
      ["Mobile", [["Impressions"]]],
    ]);
  });

  it("labels the row-totals group once, on the top header row, under several column dimensions", () => {
    const defs = getColumnDefForPivot(
      config({ colDimensionNames: ["device", "os"] }),
      { device: ["Desktop", "Mobile"], os: ["Android", "Linux"] },
      { impressions: 30, c0v0_c1v1m0: 10, c0v1_c1v0m0: 20 },
    );

    expect(headerTree(defs)).toStrictEqual([
      ["Device", [["OS", [["Country"]]]]],
      ["Total", [["", [["Impressions"]]]]],
      ["Desktop", [["Linux", [["Impressions"]]]]],
      ["Mobile", [["Android", [["Impressions"]]]]],
    ]);
  });

  it('labels the row-totals group "Grand Total" when a column value already reads "Total"', () => {
    const defs = getColumnDefForPivot(
      config({}),
      { device: ["Desktop", "Total"] },
      { impressions: 30, c0v0m0: 10, c0v1m0: 20 },
    );

    expect(headerTree(defs)).toStrictEqual([
      ["Device", [["Country"]]],
      ["Grand Total", [["Impressions"]]],
      ["Desktop", [["Impressions"]]],
      ["Total", [["Impressions"]]],
    ]);
  });

  it("leaves the row-totals group out when the totals column is hidden", () => {
    const defs = getColumnDefForPivot(
      config({ pivot: { showTotalsColumn: false } }),
      { device: ["Desktop", "Mobile"] },
      { impressions: 30, c0v0m0: 10, c0v1m0: 20 },
    );

    expect(headerTree(defs)).toStrictEqual([
      ["Device", [["Country"]]],
      ["Desktop", [["Impressions"]]],
      ["Mobile", [["Impressions"]]],
    ]);
  });

  it("adds no row-totals group to a flat table", () => {
    const defs = getColumnDefForPivot(
      config({
        isFlat: true,
        rowDimensionNames: ["country", "device"],
        colDimensionNames: [],
        pivot: {
          tableMode: "flat",
          columns: [
            { id: "country", title: "Country", type: PivotChipType.Dimension },
            { id: "device", title: "Device", type: PivotChipType.Dimension },
            {
              id: "impressions",
              title: "Impressions",
              type: PivotChipType.Measure,
            },
          ],
        },
      }),
      undefined,
      { impressions: 30 },
    );

    expect(headerTree(defs)).toStrictEqual([
      ["Country"],
      ["Device"],
      ["Impressions"],
    ]);
  });
});
