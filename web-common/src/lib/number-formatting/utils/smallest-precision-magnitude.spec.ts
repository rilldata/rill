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

  // NOTE: most digits representable in js
  [79879879710.7237, -4],
  [-79879879710.7237, -4],

  // NOTE: this number is too small to represent in js,
  // so it will be rounded to zero
  [2.2247239873252e-3523, 0],

  // Smallest positive number that can be represented in js
  [Number.MIN_VALUE, -324],
  [-Number.MIN_VALUE, -324],

  // Largest number that can be represented in js, 1.7976931348623157e+308
  [Number.MAX_VALUE, 292],
  [-Number.MAX_VALUE, 292],

  // number that can be represented in js once rounded,
  // which has more digits than can be stored
  [2.2247239873252e-308, -324],
];

describe("smallestPrecisionMagnitude", () => {
  it("NaN on non-numbers", () => {
    expect(() => smallestPrecisionMagnitude("foo" as any)).toThrow();
    expect(() => smallestPrecisionMagnitude(undefined as any)).toThrow();
  });
  testCases.forEach(([input, output]) => {
    it(`returns the order of magnitude of the most precise digit in each number ${[
      input,
      output,
    ]}`, () => {
      expect(smallestPrecisionMagnitude(input)).toBe(output);
    });
  });
});
