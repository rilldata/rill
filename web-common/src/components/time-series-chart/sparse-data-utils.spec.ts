import { describe, it, expect } from "vitest";
import {
  computeSegments,
  bridgeGaps,
  snapToNearestNonNull,
} from "./sparse-data-utils";

const identity = (d: number | null) => d;
const cloneWith = (_d: number | null, v: number): number | null => v;

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

describe("snapToNearestNonNull", () => {
  const big = 1000; // effectively unbounded snap distance

  it("snaps to the closer of two non-null points", () => {
    const data = [10, null, null, null, 20];
    // 1.4 is closer to index 0 than to index 4
    expect(snapToNearestNonNull(1.4, [data], identity, big)).toBe(0);
    // 3.0 is closer to index 4
    expect(snapToNearestNonNull(3.0, [data], identity, big)).toBe(4);
  });

  it("snaps off a null directly under the cursor to the nearest non-null", () => {
    const data = [10, null, null, null, 20];
    // Exactly on a null (index 2), equidistant: prefers the right (index 4
    // is reached before index 0 by the outward scan at equal distance)
    expect(snapToNearestNonNull(2, [data], identity, big)).toBe(4);
  });

  it("returns null when the nearest non-null is beyond maxDistance", () => {
    const data = [10, null, null, null, null, null, 20];
    // Cursor at index 3 is 3 away from both ends; maxDistance 2 excludes them
    expect(snapToNearestNonNull(3, [data], identity, 2)).toBeNull();
  });

  it("returns null for all-null data", () => {
    expect(snapToNearestNonNull(1, [[null, null, null]], identity, big)).toBe(
      null,
    );
  });

  it("returns null for empty data", () => {
    expect(snapToNearestNonNull(0, [[]], identity, big)).toBeNull();
  });

  it("snaps to a single non-null point within distance", () => {
    const data = [null, null, 42, null, null];
    expect(snapToNearestNonNull(1, [data], identity, big)).toBe(2);
    expect(snapToNearestNonNull(1, [data], identity, 0.5)).toBeNull();
  });

  it("clamps a fractional index outside the data range", () => {
    const data = [5, null, null];
    expect(snapToNearestNonNull(-3, [data], identity, big)).toBe(0);
  });

  describe("comparison (multiple series)", () => {
    it("snaps to an index where only the secondary series is non-null", () => {
      const primary = [10, null, null, null, null];
      const secondary = [null, null, null, 20, null];
      // Cursor near index 3: primary null there, secondary has a value
      expect(snapToNearestNonNull(3, [primary, secondary], identity, big)).toBe(
        3,
      );
    });

    it("snaps to whichever line's point is closest to the cursor", () => {
      const primary = [10, null, null, null, null];
      const secondary = [null, null, null, null, 20];
      // Cursor at 1.4 is closest to primary's index 0
      expect(
        snapToNearestNonNull(1.4, [primary, secondary], identity, big),
      ).toBe(0);
      // Cursor at 3.6 is closest to secondary's index 4
      expect(
        snapToNearestNonNull(3.6, [primary, secondary], identity, big),
      ).toBe(4);
    });

    it("handles series of different lengths", () => {
      const primary = [10];
      const secondary = [null, null, 20];
      expect(snapToNearestNonNull(2, [primary, secondary], identity, big)).toBe(
        2,
      );
    });
  });
});

describe("bridgeGaps", () => {
  it("bridges gaps when connectNulls is true", () => {
    const data: (number | null)[] = [10, null, 20];
    const result = bridgeGaps(data, identity, cloneWith, true);
    expect(result.values[1]).toBe(0); // filled with zero
    expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 2 }]);
  });

  it("fills multiple consecutive nulls with zeros", () => {
    const data: (number | null)[] = [3, null, null, null, 7];
    const result = bridgeGaps(data, identity, cloneWith, true);
    expect(result.values[1]).toBe(0);
    expect(result.values[2]).toBe(0);
    expect(result.values[3]).toBe(0);
    expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 4 }]);
  });

  it("does not bridge when connectNulls is false", () => {
    const data: (number | null)[] = [10, null, 20];
    const result = bridgeGaps(data, identity, cloneWith, false);
    expect(result.values[1]).toBeNull();
    expect(result.bridgedSegments).toEqual([
      { startIndex: 0, endIndex: 0 },
      { startIndex: 2, endIndex: 2 },
    ]);
  });

  it("bridges wide gaps regardless of width", () => {
    const data: (number | null)[] = [10, null, null, null, null, 20];
    const result = bridgeGaps(data, identity, cloneWith, true);
    expect(result.values).toEqual([10, 0, 0, 0, 0, 20]);
    expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 5 }]);
  });

  describe("singleton detection with connectNulls on", () => {
    it("merges all interior singletons into one continuous segment", () => {
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
      const result = bridgeGaps(data, identity, cloneWith, true);

      // Every gap bridged: one continuous segment, no singletons.
      expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 10 }]);

      const singletons = result.bridgedSegments
        .filter((s) => s.startIndex === s.endIndex)
        .map((s) => s.startIndex);
      expect(singletons).toEqual([]);
    });

    it("merges singletons separated by single-null gaps", () => {
      const data: (number | null)[] = [10, null, 20, null, 30];
      const result = bridgeGaps(data, identity, cloneWith, true);

      expect(result.bridgedSegments).toEqual([{ startIndex: 0, endIndex: 4 }]);

      const singletons = result.bridgedSegments
        .filter((s) => s.startIndex === s.endIndex)
        .map((s) => s.startIndex);
      expect(singletons).toEqual([]);
    });
  });
});
