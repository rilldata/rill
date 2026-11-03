import { SHOW_MORE_BUTTON } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import type {
  PivotDataRow,
  PivotDataState,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { describe, expect, it } from "vitest";
import { hasMoreRows } from "./table-rows";

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
