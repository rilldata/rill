import { MAX_PIVOT_COMPARISON_MEASURES } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import {
  canEnablePivotComparison,
  getSortForAccessor,
  sortAcessors,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotChipData,
  PivotDataStoreConfig,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { SortingState } from "tanstack-table-8-svelte-5";
import { describe, expect, it } from "vitest";

describe("sortAcessors function", () => {
  it("should correctly sort accessors with basic sorting", () => {
    const input = ["c1v2m3", "c0v0m0", "c2v3m1"];
    const expected = ["c0v0m0", "c1v2m3", "c2v3m1"];
    expect(sortAcessors(input)).toEqual(expected);
  });

  it("should sort accessors with varying numbers of c<num>v<num> sequences", () => {
    const input = ["c0v1_c1v2m0", "c0v0_c1v1m2", "c0v0_c1v0m1"];
    const expected = ["c0v0_c1v0m1", "c0v0_c1v1m2", "c0v1_c1v2m0"];
    expect(sortAcessors(input)).toEqual(expected);
  });

  it("should sort accessors with the same c-v values but different m values", () => {
    const input = ["c0v1m3", "c0v1m1", "c0v1m2"];
    const expected = ["c0v1m1", "c0v1m2", "c0v1m3"];
    expect(sortAcessors(input)).toEqual(expected);
  });

  it("should sort accessors with multiple c-v-m sequences, including different-number lengths", () => {
    const input = ["c1v10_c2v20m30", "c1v2_c2v3m4", "c1v10_c2v3m4"];
    const expected = ["c1v2_c2v3m4", "c1v10_c2v3m4", "c1v10_c2v20m30"];
    expect(sortAcessors(input)).toEqual(expected);
  });
});

describe("canEnablePivotComparison", () => {
  const chips = (type: PivotChipType, count: number): PivotChipData[] =>
    Array.from({ length: count }, (_, i) => ({
      id: `${type}_${i}`,
      title: `${type} ${i}`,
      type,
    }));

  const stateWithColumns = (columns: PivotChipData[]) =>
    ({ columns }) as PivotState;

  const comparisonStart = "2024-01-01T00:00:00Z";

  it("is enabled when the measure count is within the limit", () => {
    const state = stateWithColumns(
      chips(PivotChipType.Measure, MAX_PIVOT_COMPARISON_MEASURES),
    );
    expect(canEnablePivotComparison(state, comparisonStart)).toBe(true);
  });

  it("is disabled when the measure count exceeds the limit", () => {
    const state = stateWithColumns(
      chips(PivotChipType.Measure, MAX_PIVOT_COMPARISON_MEASURES + 1),
    );
    expect(canEnablePivotComparison(state, comparisonStart)).toBe(false);
  });

  it("only counts measure chips towards the limit", () => {
    const state = stateWithColumns([
      ...chips(PivotChipType.Dimension, 5),
      ...chips(PivotChipType.Time, 2),
      ...chips(PivotChipType.Measure, MAX_PIVOT_COMPARISON_MEASURES),
    ]);
    expect(canEnablePivotComparison(state, comparisonStart)).toBe(true);
  });

  it("is disabled without a comparison time range", () => {
    const state = stateWithColumns(chips(PivotChipType.Measure, 1));
    expect(canEnablePivotComparison(state, undefined)).toBe(false);
  });
});

describe("getSortForAccessor", () => {
  const timeRange = {
    start: "2024-01-01T00:00:00Z",
    end: "2024-02-01T00:00:00Z",
  };
  const columnDimensionAxes = { touched_by_name: ["alice", "bob"] };

  function config(sorting: SortingState): PivotDataStoreConfig {
    return {
      measureNames: ["touches"],
      rowDimensionNames: ["mutation_path"],
      colDimensionNames: ["touched_by_name"],
      allMeasures: [],
      allDimensions: [],
      whereFilter: createAndExpression([]),
      pivot: {
        rows: [],
        columns: [],
        sorting,
        expanded: {},
        columnPage: 1,
        rowPage: 1,
        enableComparison: true,
        tableMode: "nest",
        activeCell: null,
        showTotalsColumn: true,
        showTotalsRow: true,
      },
      time: {
        timeStart: timeRange.start,
        timeEnd: timeRange.end,
        timeZone: "UTC",
        timeDimension: "eventDate",
      },
      comparisonTime: undefined,
      enableComparison: false,
      searchText: undefined,
      isFlat: false,
    };
  }

  it("sorts the row axis by the anchor dimension", () => {
    expect(
      getSortForAccessor(
        "mutation_path",
        config([{ id: "mutation_path", desc: true }]),
        columnDimensionAxes,
      ),
    ).toEqual({
      sortPivotBy: [{ name: "mutation_path", desc: true }],
      timeRange,
    });
  });

  it("sorts by a measure's row total", () => {
    expect(
      getSortForAccessor(
        "mutation_path",
        config([{ id: "touches", desc: false }]),
        columnDimensionAxes,
      ),
    ).toEqual({
      sortPivotBy: [{ name: "touches", desc: false }],
      timeRange,
    });
  });

  it("sorts by a measure under one column value", () => {
    expect(
      getSortForAccessor(
        "mutation_path",
        config([{ id: "c0v1m0", desc: true }]),
        columnDimensionAxes,
      ),
    ).toEqual({
      where: createAndExpression([
        createInExpression("touched_by_name", ["bob"]),
      ]),
      sortPivotBy: [{ name: "touches", desc: true }],
      timeRange,
    });
  });

  // A sort outlives its field in the URL, bookmarks and reports; the time
  // dimension's own name is the common case since its column is keyed by grain.
  it("ignores a sort on a field that is not in the pivot", () => {
    expect(
      getSortForAccessor(
        "mutation_path",
        config([{ id: "eventDate", desc: true }]),
        columnDimensionAxes,
      ),
    ).toEqual({ sortPivotBy: [], timeRange });
  });

  it("ignores a sort on a row dimension that is not the anchor", () => {
    expect(
      getSortForAccessor(
        "mutation_path",
        config([{ id: "touched_by_name", desc: true }]),
        columnDimensionAxes,
      ),
    ).toEqual({ sortPivotBy: [], timeRange });
  });

  it("ignores a comparison sort while the comparison is off", () => {
    expect(
      getSortForAccessor(
        "mutation_path",
        config([{ id: "touches__delta_abs", desc: true }]),
        columnDimensionAxes,
      ),
    ).toEqual({ sortPivotBy: [], timeRange });
  });

  it("ignores an accessor beyond the pivot's columns or measures", () => {
    for (const id of ["c1v0m0", "c0v0m1"]) {
      expect(
        getSortForAccessor(
          "mutation_path",
          config([{ id, desc: true }]),
          columnDimensionAxes,
        ),
      ).toEqual({ sortPivotBy: [], timeRange });
    }
  });
});
