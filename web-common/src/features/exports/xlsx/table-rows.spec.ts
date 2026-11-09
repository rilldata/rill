import { SHOW_MORE_BUTTON } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import type {
  PivotDataRow,
  PivotDataState,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { readable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  hasMoreRows,
  MAX_FLAT_TABLE_ROWS,
  tableExportTable,
} from "./table-rows";

// Stands in for the runtime behind the table's query builder: `total` source
// rows, served by limit and offset, with every request recorded.
const source = vi.hoisted(() => ({
  total: 0,
  fail: false,
  requests: [] as { limit: number; offset: number }[],
}));

vi.mock(
  "@rilldata/web-common/features/dashboards/pivot/pivot-data-store",
  () => ({
    createTableCellQuery: (
      _ctx: unknown,
      _config: unknown,
      _columnAxes: unknown,
      _totalsRow: unknown,
      _rowValues: unknown,
      limit: string,
      offset: string,
    ) => {
      const request = { limit: Number(limit), offset: Number(offset) };
      source.requests.push(request);
      if (source.fail) {
        return readable({ isError: true, error: new Error("over the cap") });
      }
      const count = Math.max(
        0,
        Math.min(request.limit, source.total - request.offset),
      );
      return readable({
        isSuccess: true,
        data: {
          data: Array.from({ length: count }, (_, i) => ({
            country: `c${request.offset + i}`,
          })),
        },
      });
    },
  }),
);

function state(rows: PivotDataRow[], reachedEnd?: boolean): PivotDataState {
  return {
    isFetching: false,
    data: rows,
    columnDef: [],
    assembled: true,
    totalColumns: 0,
    reachedEndForRowData: reachedEnd,
  };
}

function config(rowLimit?: number): PivotDataStoreConfig {
  return {
    rowDimensionNames: ["country"],
    pivot: { rowLimit },
  } as unknown as PivotDataStoreConfig;
}

const rows = (n: number): PivotDataRow[] =>
  Array.from({ length: n }, (_, i) => ({ country: `c${i}` }));

describe("hasMoreRows", () => {
  it("is true for a pending show-more row", () => {
    expect(
      hasMoreRows(
        state([...rows(5), { country: SHOW_MORE_BUTTON }]),
        config(5),
      ),
    ).toBe(true);
  });

  it("is true at the row limit, since the store only knows about more rows through the show-more row", () => {
    expect(hasMoreRows(state(rows(100), true), config(100))).toBe(true);
    expect(hasMoreRows(state(rows(40), false), config(100))).toBe(false);
  });

  it("is true on a page boundary before the store has seen the end", () => {
    expect(hasMoreRows(state(rows(50), false), config())).toBe(true);
    expect(hasMoreRows(state(rows(50), true), config())).toBe(false);
    expect(hasMoreRows(state(rows(30), false), config())).toBe(false);
    expect(hasMoreRows(state([], false), config())).toBe(false);
  });
});

describe("tableExportTable", () => {
  const component = {
    id: "table",
    parent: { client: {} },
    specStore: readable({ metrics_view: "mv" }),
  } as unknown as BaseCanvasComponent;
  const flatConfig = {
    isFlat: true,
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    time: { timeZone: "UTC", timeDimension: "" },
    pivot: {},
  } as unknown as PivotDataStoreConfig;
  // One page on screen, with more to scroll to.
  const onScreen = state(rows(50), false);

  beforeEach(() => {
    source.total = 0;
    source.fail = false;
    source.requests = [];
  });

  it("never asks the runtime for more rows than its limit", async () => {
    source.total = 12_345;
    const table = await tableExportTable(component, onScreen, flatConfig);
    expect(table.rows).toHaveLength(MAX_FLAT_TABLE_ROWS);
    expect(table.moreRows).toBe(true);
    expect(source.requests).toStrictEqual([
      { limit: MAX_FLAT_TABLE_ROWS, offset: 0 },
      { limit: 1, offset: MAX_FLAT_TABLE_ROWS },
    ]);
  });

  it("knows a table of exactly the cap is complete", async () => {
    source.total = MAX_FLAT_TABLE_ROWS;
    const table = await tableExportTable(component, onScreen, flatConfig);
    expect(table.rows).toHaveLength(MAX_FLAT_TABLE_ROWS);
    expect(table.moreRows).toBe(false);
  });

  it("loads every row of a table below the cap in one request", async () => {
    source.total = 120;
    const table = await tableExportTable(component, onScreen, flatConfig);
    expect(table.rows).toHaveLength(120);
    expect(table.moreRows).toBe(false);
    expect(source.requests).toHaveLength(1);
  });

  it("falls back to the rows on screen when the rows can't be loaded", async () => {
    source.fail = true;
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    const table = await tableExportTable(component, onScreen, flatConfig);
    expect(table.rows).toHaveLength(50);
    expect(table.moreRows).toBe(true);
    expect(warn).toHaveBeenCalledOnce();
    warn.mockRestore();
  });
});
