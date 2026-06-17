import { describe, it, expect } from "vitest";
import { computeSegments, bridgeGaps } from "./sparse-data-utils";

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
