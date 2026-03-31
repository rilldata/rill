import { describe, it, expect } from "vitest";
import { computeSegments, bridgeSmallGaps } from "./sparse-data-utils";

const identity = (d: number | null) => d;
const cloneWith = (_d: number | null, v: number): number | null => v;
// 1:1 pixel mapping so gap width = index difference
const xPixel = (i: number) => i;

describe("computeSegments", () => {
  it("finds contiguous non-null segments", () => {
    const data = [1, null, null, 2, 3, null, 4];
    expect(computeSegments(data, identity)).toEqual([
      { startIndex: 0, endIndex: 0 },
      { startIndex: 3, endIndex: 4 },
      { startIndex: 6, endIndex: 6 },
    ]);
  });

  it("returns empty for all-null data", () => {
    expect(computeSegments([null, null], identity)).toEqual([]);
  });

  it("returns single segment for all-non-null data", () => {
    expect(computeSegments([1, 2, 3], identity)).toEqual([
      { startIndex: 0, endIndex: 2 },
    ]);
  });
});

describe("bridgeSmallGaps", () => {
  it("bridges small gaps when connectNulls is true", () => {
    // Gap of 2 indices (< default 36px threshold with 1:1 pixel mapping)
    const data: (number | null)[] = [10, null, 20];
    const result = bridgeSmallGaps(data, identity, cloneWith, xPixel, true);
    expect(result.values[1]).toBe(15); // linearly interpolated
    expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 2 }]);
  });

  it("does not bridge when connectNulls is false", () => {
    const data: (number | null)[] = [10, null, 20];
    const result = bridgeSmallGaps(data, identity, cloneWith, xPixel, false);
    expect(result.values[1]).toBeNull();
    expect(result.bridgedSegments).toEqual([
      { startIndex: 0, endIndex: 0 },
      { startIndex: 2, endIndex: 2 },
    ]);
  });

  it("does not bridge gaps wider than maxGapPx", () => {
    // With 1:1 pixel mapping and maxGapPx=2, a gap of 3 indices won't bridge
    const data: (number | null)[] = [10, null, null, 20];
    const result = bridgeSmallGaps(data, identity, cloneWith, xPixel, true, 2);
    expect(result.values[1]).toBeNull();
    expect(result.values[2]).toBeNull();
    expect(result.bridgedSegments).toEqual([
      { startIndex: 0, endIndex: 0 },
      { startIndex: 3, endIndex: 3 },
    ]);
  });

  describe("singleton detection with connectNulls on", () => {
    it("singleton surrounded by wide gaps remains a singleton in bridgedSegments", () => {
      // Gap too wide to bridge (> maxGapPx=2): singleton at index 5 stays isolated
      const data: (number | null)[] = [
        10,
        null,
        null,
        null,
        null,
        50,
        null,
        null,
        null,
        null,
        100,
      ];
      const result = bridgeSmallGaps(
        data,
        identity,
        cloneWith,
        xPixel,
        true,
        2,
      );

      // The singleton at index 5 should still appear as a singleton segment
      const singletons = result.bridgedSegments
        .filter((s) => s.startIndex === s.endIndex)
        .map((s) => s.startIndex);

      expect(singletons).toContain(0);
      expect(singletons).toContain(5);
      expect(singletons).toContain(10);
    });

    it("singleton gets merged when gap is small enough to bridge", () => {
      // Gap of 2 (within maxGapPx=36): singleton should get bridged into neighbors
      const data: (number | null)[] = [10, null, 20, null, 30];
      const result = bridgeSmallGaps(data, identity, cloneWith, xPixel, true);

      // All gaps bridged: single continuous segment
      expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 4 }]);

      const singletons = result.bridgedSegments
        .filter((s) => s.startIndex === s.endIndex)
        .map((s) => s.startIndex);
      expect(singletons).toEqual([]);
    });
  });
});
