import type { NumberParts } from "../humanizer-types";
import {
  formatNumWithOrderOfMag,
  orderOfMagnitudeEng,
} from "./format-with-order-of-magnitude";
import { describe, it, expect } from "vitest";

type TestArgs = [number, number, number, boolean, boolean?, boolean?];

const testCases: [TestArgs, NumberParts][] = [
  [[Infinity, 3, 4, true], { int: "∞", dot: "", frac: "", suffix: "" }],
  [
    [-Infinity, 3, 4, true],
    { neg: "-", int: "∞", dot: "", frac: "", suffix: "" },
  ],
  [[NaN, 3, 4, true], { int: "NaN", dot: "", frac: "", suffix: "" }],
  [[0, 5, 4, false], { int: "0", dot: ".", frac: "", suffix: "E5" }],
  [[0, 5, 4, true], { int: "0", dot: ".", frac: "0000", suffix: "E5" }],
  [[0, -5, 2, true], { int: "0", dot: ".", frac: "00", suffix: "E-5" }],

  [[1, 3, 5, false], { int: "0", dot: ".", frac: "001", suffix: "E3" }],
  [[1, 3, 5, true], { int: "0", dot: ".", frac: "00100", suffix: "E3" }],

  //  stripCommas = true
  [
    [1, -3, 5, false, false, true],
    { int: "1000", dot: "", frac: "", suffix: "E-3" },
  ],
  [
    [1, -3, 5, true, false, true],
    { int: "1000", dot: ".", frac: "00000", suffix: "E-3" },
  ],

  // stripCommas = false (by default)
  [[1, -3, 5, false], { int: "1,000", dot: "", frac: "", suffix: "E-3" }],
  [[1, -3, 5, true], { int: "1,000", dot: ".", frac: "00000", suffix: "E-3" }],

  [[0.001, 0, 5, false], { int: "0", dot: ".", frac: "001", suffix: "E0" }],
  [[0.001, 0, 5, true], { int: "0", dot: ".", frac: "00100", suffix: "E0" }],

  [[0.001, -3, 5, false], { int: "1", dot: "", frac: "", suffix: "E-3" }],
  [[0.001, -3, 5, true], { int: "1", dot: ".", frac: "00000", suffix: "E-3" }],

  [
    [710.7237956, 0, 5, true],
    { int: "710", dot: ".", frac: "72380", suffix: "E0" },
  ],
  [
    [710.7237956, 0, 5, false],
    { int: "710", dot: ".", frac: "72380", suffix: "E0" },
  ],

  // yes trailing dot
  [
    [710.272337956, 0, 0, true, true],
    { int: "710", dot: ".", frac: "", suffix: "E0" },
  ],
  [
    [710.272337956, 0, 0, false, true],
    { int: "710", dot: ".", frac: "", suffix: "E0" },
  ],

  // no trailing dot
  [
    [710.272337956, 0, 0, true, false],
    { int: "710", dot: "", frac: "", suffix: "E0" },
  ],
  [
    [710.272337956, 0, 0, false, false],
    { int: "710", dot: "", frac: "", suffix: "E0" },
  ],

  [
    [710.7237956, 0, 2, true],
    { int: "710", dot: ".", frac: "72", suffix: "E0" },
  ],
  [
    [710.7237956, 0, 2, false],
    { int: "710", dot: ".", frac: "72", suffix: "E0" },
  ],

  // not stripping commas
  [
    [523523710.7237956, 0, 5, true],
    { int: "523,523,710", dot: ".", frac: "72380", suffix: "E0" },
  ],
  [
    [523523710.7237956, 0, 5, false],
    { int: "523,523,710", dot: ".", frac: "72380", suffix: "E0" },
  ],
  // yes stripping commas
  [
    [523523710.7237956, 0, 5, true, false, true],
    { int: "523523710", dot: ".", frac: "72380", suffix: "E0" },
  ],
  [
    [523523710.7237956, 0, 5, false, false, true],
    { int: "523523710", dot: ".", frac: "72380", suffix: "E0" },
  ],

  [
    [0.00087000001, -3, 5, false],
    { int: "0", dot: ".", frac: "87000", suffix: "E-3" },
  ],
  [
    [0.00087000001, -3, 5, true],
    { int: "0", dot: ".", frac: "87000", suffix: "E-3" },
  ],

  [[0.00087, -3, 5, false], { int: "0", dot: ".", frac: "87", suffix: "E-3" }],
  [
    [0.00087, -3, 5, true],
    { int: "0", dot: ".", frac: "87000", suffix: "E-3" },
  ],

  // same as above but negative
  [
    [-0.00087000001, -3, 5, false],
    { neg: "-", int: "0", dot: ".", frac: "87000", suffix: "E-3" },
  ],
  [
    [-0.00087000001, -3, 5, true],
    { neg: "-", int: "0", dot: ".", frac: "87000", suffix: "E-3" },
  ],

  [
    [-0.00087, -3, 5, false],
    { neg: "-", int: "0", dot: ".", frac: "87", suffix: "E-3" },
  ],
  [
    [-0.00087, -3, 5, true],
    { neg: "-", int: "0", dot: ".", frac: "87000", suffix: "E-3" },
  ],
];

const integerTestCases: [TestArgs, NumberParts][] = [
  [[710, 0, 0, true, true], { int: "710", dot: "", frac: "", suffix: "E0" }],
  [[710, 0, 0, false, true], { int: "710", dot: "", frac: "", suffix: "E0" }],

  [
    [710, 0, 5, true, true],
    { int: "710", dot: ".", frac: "00000", suffix: "E0" },
  ],
  [[710, 0, 5, false, true], { int: "710", dot: "", frac: "", suffix: "E0" }],
];

describe("formatNumWithOrderOfMag", () => {
  it("throw on non-numbers", () => {
    expect(() => formatNumWithOrderOfMag("asdf" as any, -3, 5, true)).toThrow();
    expect(() => formatNumWithOrderOfMag(undefined, -3, 5, true)).toThrow();
  });

  testCases.forEach(([input, output]) => {
    it(`returns the correct split string in case: ${input}`, () => {
      expect(formatNumWithOrderOfMag(...input)).toEqual(output);
    });
  });

  integerTestCases.forEach(([input, output]) => {
    it(`should have correct "dot" and padding for int input targeting e0: ${input}`, () => {
      expect(formatNumWithOrderOfMag(...input)).toEqual(output);
    });
  });
});

describe("orderOfMagnitudeEng", () => {
  [
    [0, 0],
    [2.23434, 0],
    [10, 0],
    [210, 0],
    [3210, 3],
    [43210, 3],
    [9_876_543_210, 9],
    [876_543_210, 6],
    [76_543_210, 6],
    [6_543_210, 6],
    [0.1, -3],
    [0.01, -3],
    [0.001, -3],
    [0.000_000_000_001, -12],
    [0.000_000_000_01, -12],
    [0.000_000_000_1, -12],
  ].forEach(([input, output]) => {
    it(`returns the correct engineering order of magnitude: ${input}`, () => {
      expect(orderOfMagnitudeEng(input)).toEqual(output);
    });
  });
});
