import { FormatterFactoryOptions, NumberKind } from "../humanizer-types";
import {
  SingleDigitTimesPowerOfTenFormatter,
  closeToIntTimesPowerOfTen,
} from "./SingleDigitTimesPowerOfTen";
import { describe, it, expect } from "vitest";

const baseOptions: FormatterFactoryOptions = {
  strategy: "singleDigitTimesPowerOfTen",
  padWithInsignificantZeros: true,
  numberKind: NumberKind.ANY,
  onInvalidInput: "doNothing",
};

const closeToIntTimesPowerOfTenCases: [number, boolean][] = [
  [0.00009999999999999, true],
  [0.00019999999999999, true],
  [0.00039999999999999, true],
  [0.00000999999999999, true],
  [0.9999999999999999, true],
  [0.0030000000003, true],

  [0, true],
  [0, true],
  [0, true],

  [1, true],
  [1, true],
  [1, true],

  [30_000_000, true],
  [30_000_000, true],
  [30_000_000, true],

  [10_000, true],
  [10_000, true],
  [10_000, true],

  [10, true],
  [10, true],
  [10, true],

  [0.005, true],
  [0.005, true],
  [0.005, true],

  [0.000_000_200, true],
  [0.000_000_200, true],
  [0.000_000_200, true],

  [12_320_000, false],
  [12_320_000, false],
  [12_320_000, false],

  [12_000, false],
  [12_000, false],
  [12_000, false],

  [12_320, false],
  [12_320, false],
  [12_320, false],
  [12.23, false],
  [12.23, false],
  [12.23, false],

  [0.001432, false],
  [0.001423, false],
  [0.001423, false],

  [0.000_000_234_32, false],
  [0.000_000_234_32, false],
  [0.000_000_234_32, false],
];

describe("closeToIntTimesPowerOfTen correctly detects whether numbers are close to a single digit multiple of a power of 10", () => {
  closeToIntTimesPowerOfTenCases.forEach(([input, output]) => {
    it(`closeToIntTimesPowerOfTen correct for: ${input}`, () => {
      expect(closeToIntTimesPowerOfTen(input)).toEqual(output);
    });
  });
});

const testCases: [
  number,
  {
    maxDigitsRightSmallNums?: number;
    maxDigitsRightSuffixNums?: number;
    padWithInsignificantZeros?: boolean;
    numberKind?: NumberKind;
  },
  string
][] = [
  [0, {}, "0"],
  [0, { numberKind: NumberKind.DOLLAR }, "$0"],
  [0, { numberKind: NumberKind.PERCENT }, "0%"],

  [1, {}, "1"],
  [1, { numberKind: NumberKind.DOLLAR }, "$1"],
  [1, { numberKind: NumberKind.PERCENT }, "100%"],

  [30_000_000, {}, "30M"],
  [30_000_000, { numberKind: NumberKind.DOLLAR }, "$30M"],
  [30_000_000, { numberKind: NumberKind.PERCENT }, "3B%"],

  [10_000, {}, "10k"],
  [10_000, { numberKind: NumberKind.DOLLAR }, "$10k"],
  [10_000, { numberKind: NumberKind.PERCENT }, "1M%"],

  [10, {}, "10"],
  [10, { numberKind: NumberKind.DOLLAR }, "$10"],
  [10, { numberKind: NumberKind.PERCENT }, "1k%"],

  [0.005, {}, "5e-3"],
  [0.005, { numberKind: NumberKind.DOLLAR }, "$5e-3"],
  [0.005, { numberKind: NumberKind.PERCENT }, "500e-3%"],

  [0.000_000_200, {}, "200e-9"],
  [0.000_000_200, { numberKind: NumberKind.DOLLAR }, "$200e-9"],
  [0.000_000_200, { numberKind: NumberKind.PERCENT }, "20e-6%"],
];

describe("default formatter, default options `.stringFormat()`", () => {
  testCases.forEach(([input, options, output]) => {
    it(`returns the correct split string in case: ${input}`, () => {
      const formatter = new SingleDigitTimesPowerOfTenFormatter([input], {
        ...baseOptions,
        ...options,
      });
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});

// const errorCases: [
//   number,
//   {
//     maxDigitsRightSmallNums?: number;
//     maxDigitsRightSuffixNums?: number;
//     padWithInsignificantZeros?: boolean;
//     numberKind?: NumberKind;
//   }
// ][] = [
//   [12_320_000, {}],
//   [12_320_000, { numberKind: NumberKind.DOLLAR }],
//   [12_320_000, { numberKind: NumberKind.PERCENT }],

//   [12_000, {}],
//   [12_000, { numberKind: NumberKind.DOLLAR }],
//   [12_000, { numberKind: NumberKind.PERCENT }],

//   [12_320, {}],
//   [12_320, { numberKind: NumberKind.DOLLAR }],
//   [12_320, { numberKind: NumberKind.PERCENT }],
//   [12.23, {}],
//   [12.23, { numberKind: NumberKind.DOLLAR }],
//   [12.23, { numberKind: NumberKind.PERCENT }],

//   [0.001432, {}],
//   [0.001423, { numberKind: NumberKind.DOLLAR }],
//   [0.001423, { numberKind: NumberKind.PERCENT }],

//   [0.000_000_234_32, {}],
//   [0.000_000_234_32, { numberKind: NumberKind.DOLLAR }],
//   [0.000_000_234_32, { numberKind: NumberKind.PERCENT }],
// ];

//FIXME re-enable this test when we have a better way to handle invalid inputs
// describe("SingleDigitTimesPowerOfTenFormatter, returns empty NumberParts on invalid inputs", () => {
//   errorCases.forEach(([input, options]) => {
//     it(`throws an error for input: ${input}`, () => {
//       const formatter = new SingleDigitTimesPowerOfTenFormatter([input], {
//         ...baseOptions,
//         ...options,
//         ...{ onInvalidInput: "consoleWarn" },
//       });
//       expect(() => formatter.stringFormat(input)).toEqual("");
//     });
//   });
// });

const closeCases: [number, string][] = [
  [0.00009999999999999, "100e-6"],
  [0.00019999999999999, "200e-6"],
  [0.00039999999999999, "400e-6"],
  [0.00000999999999999, "10e-6"],
  [0.9999999999999999, "1"],
  [0.0030000000003, "3e-3"],
];

describe("SingleDigitTimesPowerOfTenFormatter handles cases within an rounding error", () => {
  closeCases.forEach(([input, output]) => {
    it(`returns the correct split string in case: ${input}, and does not throw an error`, () => {
      const formatter = new SingleDigitTimesPowerOfTenFormatter([input], {
        ...baseOptions,
        ...{ onInvalidInput: "throw" },
      });
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});
