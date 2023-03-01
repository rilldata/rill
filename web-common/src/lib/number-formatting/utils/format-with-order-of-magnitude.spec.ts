import type { NumberStringParts } from "../humanizer-types";
import { formatNumWithOrderOfMag } from "./format-with-order-of-magnitude";

type TestArgs = [number, number, number, boolean, boolean?];

const testCases: [TestArgs, NumberStringParts][] = [
  [[Infinity, 3, 4, true], { int: "∞", dot: "", frac: "", suffix: "" }],
  [[-Infinity, 3, 4, true], { int: "-∞", dot: "", frac: "", suffix: "" }],
  [[NaN, 3, 4, true], { int: "NaN", dot: "", frac: "", suffix: "" }],
  [[0, 5, 4, false], { int: "0", dot: ".", frac: "", suffix: "E5" }],
  [[0, 5, 4, true], { int: "0", dot: ".", frac: "0000", suffix: "E5" }],
  [[0, -5, 2, true], { int: "0", dot: ".", frac: "00", suffix: "E-5" }],

  [[1, 3, 5, false], { int: "0", dot: ".", frac: "001", suffix: "E3" }],
  [[1, 3, 5, true], { int: "0", dot: ".", frac: "00100", suffix: "E3" }],

  [[1, -3, 5, false], { int: "1000", dot: "", frac: "", suffix: "E-3" }],
  [[1, -3, 5, true], { int: "1000", dot: ".", frac: "00000", suffix: "E-3" }],

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

  [
    [523523710.7237956, 0, 5, true],
    { int: "523523710", dot: ".", frac: "72380", suffix: "E0" },
  ],
  [
    [523523710.7237956, 0, 5, false],
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
    { int: "-0", dot: ".", frac: "87000", suffix: "E-3" },
  ],
  [
    [-0.00087000001, -3, 5, true],
    { int: "-0", dot: ".", frac: "87000", suffix: "E-3" },
  ],

  [
    [-0.00087, -3, 5, false],
    { int: "-0", dot: ".", frac: "87", suffix: "E-3" },
  ],
  [
    [-0.00087, -3, 5, true],
    { int: "-0", dot: ".", frac: "87000", suffix: "E-3" },
  ],
];

const integerTestCases: [TestArgs, NumberStringParts][] = [
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
