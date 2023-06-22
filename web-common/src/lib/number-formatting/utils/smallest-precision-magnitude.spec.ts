import { smallestPrecisionMagnitude } from "./smallest-precision-magnitude";
import { describe, it, expect } from "vitest";

const testCases: [number, number][] = [
  [Infinity, NaN],
  [-Infinity, NaN],
  [NaN, NaN],
  [0, 0],
  [-0, 0],
  [0.1, -1],
  [0.000000001, -9],
  [0.000027391, -9],
  [0.973427391, -9],
  [123, 0],
  [1230, 1],
  [123000000000, 9],
  [1239347293742974, 0],
  [6000000000, 9],
  [-0.1, -1],
  [-0.000000001, -9],
  [-0.000027391, -9],
  [-0.973427391, -9],
  [-123, 0],
  [-1230, 1],
  [-123000000000, 9],
  [-1239347293742974, 0],
  [-6000000000, 9],
  [710.7237956, -7],
  [-710.7237956, -7],

  [79879879710.7237, -4], // NOTE: most digits representable in js
  [-79879879710.7237, -4], // NOTE: most digits representable in js
];

describe("smallestPrecisionMagnitude", () => {
  it("NaN on non-numbers", () => {
    expect(() => smallestPrecisionMagnitude("foo" as any)).toThrow();
    expect(() => smallestPrecisionMagnitude(undefined as any)).toThrow();
  });
  it("returns the order of magnitude of the most precise digit in each number", () => {
    testCases.forEach(([input, output]) => {
      expect(smallestPrecisionMagnitude(input)).toBe(output);
    });
  });
});
