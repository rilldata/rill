import type {
  PivotDataRow,
  PivotDataState,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SHOW_MORE_BUTTON } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type { ColumnDef } from "tanstack-table-8-svelte-5";
import { describe, expect, it } from "vitest";
import { flattenPivotData, leafColumns, visibleRows } from "./pivot-flatten";

function config(
  overrides: Partial<PivotDataStoreConfig>,
): PivotDataStoreConfig {
  return {
    measureNames: ["impressions", "ctr"],
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    allMeasures: [
      {
        name: "impressions",
        displayName: "Impressions",
        formatPreset: "humanize",
      },
      { name: "ctr", displayName: "CTR", formatPreset: "percentage" },
    ],
    allDimensions: [{ name: "country", displayName: "Country" }],
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
      color: "",
      showTotalsRow: false,
      totalsRowPosition: "top",
    } as unknown as PivotDataStoreConfig["pivot"],
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
    ...overrides,
  };
}

function state(
  columnDef: ColumnDef<PivotDataRow>[],
  data: PivotDataRow[],
  extra: Partial<PivotDataState> = {},
): PivotDataState {
  return {
    isFetching: false,
    data,
    columnDef,
    assembled: true,
    totalColumns: columnDef.length,
    ...extra,
  };
}

describe("leafColumns", () => {
  it("folds group headers into the leaf header and names the row-totals group", () => {
    const defs = [
      { id: "country", accessorFn: () => null, header: "Country" },
      {
        id: "device",
        header: "",
        columns: [{ accessorKey: "impressions", header: "Impressions" }],
      },
      {
        header: "Mobile",
        columns: [{ accessorKey: "mobile_impressions", header: "Impressions" }],
      },
    ] as ColumnDef<PivotDataRow>[];

    expect(leafColumns(defs)).toStrictEqual([
      { key: "country", groups: [], header: "Country", measure: undefined },
      {
        key: "impressions",
        groups: ["Total"],
        header: "Impressions",
        measure: undefined,
      },
      {
        key: "mobile_impressions",
        groups: ["Mobile"],
        header: "Impressions",
        measure: undefined,
      },
    ]);
  });
});

describe("visibleRows", () => {
  it("includes sub-rows only under expanded rows and skips placeholders", () => {
    const rows: PivotDataRow[] = [
      {
        country: "US",
        subRows: [{ country: "CA-state" }, { country: SHOW_MORE_BUTTON }],
      },
      { country: "IN", subRows: [{ country: "KA" }] },
    ];
    const result = visibleRows(rows, { "0": true });
    expect(result.map((r) => [r.row.country, r.depth])).toStrictEqual([
      ["US", 0],
      ["CA-state", 1],
      ["IN", 0],
    ]);
  });
});

describe("flattenPivotData", () => {
  const defs = [
    { id: "country", accessorFn: () => null, header: "Country" },
    { accessorKey: "impressions", header: "Impressions" },
    { accessorKey: "ctr", header: "CTR" },
    { accessorKey: "impressions__delta_rel", header: "Impressions Δ%" },
  ] as ColumnDef<PivotDataRow>[];

  it("types measures by their format and keeps the totals row apart from the rows", () => {
    const table = flattenPivotData(
      state(
        defs,
        [
          {
            country: "US",
            impressions: 10,
            ctr: 0.5,
            impressions__delta_rel: 0.1,
          },
        ],
        {
          totalsRowData: {
            impressions: 10,
            ctr: 0.5,
            impressions__delta_rel: 0.1,
          },
        },
      ),
      config({
        pivot: {
          expanded: {},
          showTotalsRow: true,
          totalsRowPosition: "bottom",
        } as unknown as PivotDataStoreConfig["pivot"],
      }),
    );
    expect(
      table.columns.map((c) => [c.name, c.type, c.numberFormat]),
    ).toStrictEqual([
      ["country", WorkbookColumnType.STRING, undefined],
      ["impressions", WorkbookColumnType.NUMBER, "#,##0"],
      ["ctr", WorkbookColumnType.NUMBER, "0.0%"],
      ["impressions__delta_rel", WorkbookColumnType.NUMBER, "0.0%"],
    ]);
    // A comparison column's own header leans on the column beside it; the sheet names the measure.
    expect(table.columns[3].displayName).toBe("Impressions % change");
    expect(table.rows).toStrictEqual([["US", 10, 0.5, 0.1]]);
    expect(table.totals).toStrictEqual(["Total", 10, 0.5, 0.1]);
  });

  it("types the measure cells under column dimensions by the measure the column shows", () => {
    // Under a column dimension the cell keys ("c0v0m0") no longer name the measure; the leaf's `name` does.
    const nestedDefs = [
      { id: "country", accessorFn: () => null, header: "Country" },
      {
        header: "Mobile",
        columns: [
          { accessorKey: "c0v0m0", header: "Impressions", name: "impressions" },
          {
            accessorKey: "c0v0m1",
            header: "Δ",
            name: "impressions__delta_abs",
          },
          { accessorKey: "c0v0m2", header: "CTR", name: "ctr" },
        ],
      },
    ] as unknown as ColumnDef<PivotDataRow>[];
    const table = flattenPivotData(
      state(nestedDefs, [
        { country: "US", c0v0m0: 117.50599999999999, c0v0m1: -3, c0v0m2: 0.5 },
      ]),
      config({ colDimensionNames: ["device"] }),
    );
    expect(
      table.columns.map((c) => [c.name, c.displayName, c.type, c.numberFormat]),
    ).toStrictEqual([
      ["country", "Country", WorkbookColumnType.STRING, undefined],
      ["c0v0m0", "Mobile / Impressions", WorkbookColumnType.NUMBER, "#,##0"],
      [
        "c0v0m1",
        "Mobile / Impressions change",
        WorkbookColumnType.NUMBER,
        "#,##0",
      ],
      ["c0v0m2", "Mobile / CTR", WorkbookColumnType.NUMBER, "0.0%"],
    ]);
    expect(table.rows).toStrictEqual([["US", 117.50599999999999, -3, 0.5]]);
    expect(table.totals).toBeUndefined();
  });

  it("labels nested time row headers and types flat time columns", () => {
    const nested = flattenPivotData(
      state(defs.slice(0, 2), [
        {
          country: "US",
          impressions: 10,
          subRows: [{ country: "2026-03-01T00:00:00Z", impressions: 4 }],
        },
      ]),
      config({
        rowDimensionNames: ["country", "timestamp_rill_TIME_GRAIN_MONTH"],
        pivot: { expanded: true } as unknown as PivotDataStoreConfig["pivot"],
      }),
    );
    expect(nested.rows).toStrictEqual([
      ["US", 10],
      ["Mar 2026", 4],
    ]);

    const flat = flattenPivotData(
      state(
        [
          {
            id: "timestamp_rill_TIME_GRAIN_DAY",
            accessorFn: () => null,
            header: "Day",
          },
          { accessorKey: "impressions", header: "Impressions" },
        ] as ColumnDef<PivotDataRow>[],
        [
          {
            timestamp_rill_TIME_GRAIN_DAY: "2026-03-01T00:00:00Z",
            impressions: 4,
          },
        ],
      ),
      config({
        isFlat: true,
        rowDimensionNames: ["timestamp_rill_TIME_GRAIN_DAY"],
        time: {
          timeStart: undefined,
          timeEnd: undefined,
          timeZone: "Asia/Kolkata",
          timeDimension: "timestamp",
        },
      }),
    );
    expect(flat.columns[0]).toStrictEqual({
      name: "timestamp_rill_TIME_GRAIN_DAY",
      displayName: "Day",
      type: WorkbookColumnType.DATE,
      numberFormat: "yyyy-mm-dd",
    });
    expect(flat.rows).toStrictEqual([["2026-03-01T05:30:00", 4]]);
  });
});
