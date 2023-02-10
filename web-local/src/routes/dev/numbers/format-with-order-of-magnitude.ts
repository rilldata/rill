import type { NumberStringParts } from "./number-to-string-formatters";
import { smallestPrecisionMagnitude } from "./smallest-precision-magnitude";

import _ from "lodash";

export const formatNumWithOrderOfMag2 = (
  x: number,
  newOrder: number,
  fractionDigits: number,
  padInsignificantZeros = false
): NumberStringParts => {
  if (x === Infinity) return { int: "∞", dot: "", frac: "", suffix: "" };
  if (x === -Infinity) return { int: "-∞", dot: "", frac: "", suffix: "" };
  if (Number.isNaN(x)) return { int: "NaN", dot: "", frac: "", suffix: "" };

  const suffix = "E" + newOrder;
  const dot: "." = ".";

  if (x === 0)
    return {
      int: "0",
      dot,
      frac: padInsignificantZeros ? "0".repeat(fractionDigits) : "",
      suffix,
    };

  if (padInsignificantZeros === false) {
    // get the OoM of the smallest digit
    const spm = smallestPrecisionMagnitude(x);
    // calculate the order of mag of the smallest precision digit
    // when represented in this new magnitude
    const newSpm = spm - newOrder;
    // Use the new order to set the final value of fractionDigits
    // to be used in the formatter.
    // Note that the the number of fractional digits to keep will be positive
    // only if the smallest precision OoM is negative.
    // Finally, if not using zero padding, we want to use the smaller of
    // (a) the number of digits needed for this smallest precision, or
    // (b) the target number of fraction digits

    // if padding would *add* zeros that are not needed,

    if (newSpm < 0) {
      fractionDigits = Math.min(-newSpm, fractionDigits);
    } else {
      fractionDigits = 0;
    }
    // if (spm < -5) {
    // console.log({ x, spm, newSpm, fractionDigits });
    // }
  }

  const [int, frac] = Intl.NumberFormat("en-US", {
    maximumFractionDigits: fractionDigits,
    minimumFractionDigits: fractionDigits,
  })
    .format(x / 10 ** newOrder)
    .replace(/,/g, "")
    .split(".");

  const splitStr = { int, dot, frac: frac ?? "", suffix };

  return splitStr;
};

const testCases: [[number, number, number, boolean], NumberStringParts][] = [
  [[Infinity, 3, 4, true], { int: "∞", dot: "", frac: "", suffix: "" }],
  [[-Infinity, 3, 4, true], { int: "-∞", dot: "", frac: "", suffix: "" }],
  [[NaN, 3, 4, true], { int: "NaN", dot: "", frac: "", suffix: "" }],
  [[0, 5, 4, false], { int: "0", dot: ".", frac: "", suffix: "E5" }],
  [[0, 5, 4, true], { int: "0", dot: ".", frac: "0000", suffix: "E5" }],
  [[0, -5, 2, true], { int: "0", dot: ".", frac: "00", suffix: "E-5" }],

  [[1, 3, 5, false], { int: "0", dot: ".", frac: "001", suffix: "E3" }],
  [[1, 3, 5, true], { int: "0", dot: ".", frac: "00100", suffix: "E3" }],

  [[1, -3, 5, false], { int: "1000", dot: ".", frac: "", suffix: "E-3" }],
  [[1, -3, 5, true], { int: "1000", dot: ".", frac: "00000", suffix: "E-3" }],

  [[0.001, 0, 5, false], { int: "0", dot: ".", frac: "001", suffix: "E0" }],
  [[0.001, 0, 5, true], { int: "0", dot: ".", frac: "00100", suffix: "E0" }],

  [[0.001, -3, 5, false], { int: "1", dot: ".", frac: "", suffix: "E-3" }],
  [[0.001, -3, 5, true], { int: "1", dot: ".", frac: "00000", suffix: "E-3" }],

  [
    [710.7237956, 0, 5, true],
    { int: "710", dot: ".", frac: "72380", suffix: "E0" },
  ],
  [
    [710.7237956, 0, 5, false],
    { int: "710", dot: ".", frac: "72380", suffix: "E0" },
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

export const runTest_formatNumWithOrderOfMag2 = () => {
  const caseResults = testCases.map((tc) => {
    const out = formatNumWithOrderOfMag2(...tc[0]);
    const correct = _.isEqual(out, tc[1]);
    return [correct ? "pass" : "fail", tc, out];
  });

  const overallPass = caseResults.reduce(
    (prev, tc) => prev && tc[0] === "pass",
    true
  );

  if (!overallPass) {
    console.error("runTest_formatNumWithOrderOfMag2 test FAIL:", caseResults);
    caseResults.forEach((cr) => {
      if (cr[0] === "fail") {
        console.log("---    in:", cr[1][0]);
        console.log("   target:", cr[1][1]);
        console.log("      out:", cr[2]);
      }
    });
  } else {
    console.log("runTest_formatNumWithOrderOfMag2 test PASS:", caseResults);
  }
};
