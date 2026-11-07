import { describe, expect, it } from "vitest";
import { buildChartSeries } from "./chart-series";
import type { DimensionSeriesData } from "./types";

function dimensionSeries(dimensionValue: string | null): DimensionSeriesData {
  return { dimensionValue, color: "", data: [], isFetching: false };
}

describe("buildChartSeries", () => {
  it("gives every dimension series a unique id when values repeat or stringify alike", () => {
    // A selection of `0, null, 1` keyed the null value by its index, which is also the key of `1`.
    // A selection can also hold a value twice, as a number and as a string.
    const dimData = [0, null, 1, 9304, "9304", "9304"].map((value) =>
      dimensionSeries(value as string | null),
    );

    for (const showComparison of [false, true]) {
      const ids = buildChartSeries([], dimData, showComparison).map(
        (s) => s.id,
      );
      expect(ids).toHaveLength(showComparison ? 12 : 6);
      expect(new Set(ids).size).toBe(ids.length);
    }
  });
});
