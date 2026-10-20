import { MAX_PIVOT_COMPARISON_MEASURES } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import {
  canEnablePivotComparison,
  sortAcessors,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotChipData,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
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
