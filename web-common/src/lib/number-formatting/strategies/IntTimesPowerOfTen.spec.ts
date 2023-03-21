import { FormatterFactoryOptions, NumberKind } from "../humanizer-types";
import { IntTimesPowerOfTenFormatter } from "./IntTimesPowerOfTen";

const baseOptions: FormatterFactoryOptions = {
  strategy: "intTimesPowerOfTen",
  padWithInsignificantZeros: true,
  numberKind: NumberKind.ANY,
  onInvalidInput: "doNothing",
};

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
      const formatter = new IntTimesPowerOfTenFormatter([input], {
        ...baseOptions,
        ...options,
      });
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});

const errorCases: [
  number,
  {
    maxDigitsRightSmallNums?: number;
    maxDigitsRightSuffixNums?: number;
    padWithInsignificantZeros?: boolean;
    numberKind?: NumberKind;
  }
][] = [
  [12_320_000, {}],
  [12_320_000, { numberKind: NumberKind.DOLLAR }],
  [12_320_000, { numberKind: NumberKind.PERCENT }],

  [12_000, {}],
  [12_000, { numberKind: NumberKind.DOLLAR }],
  [12_000, { numberKind: NumberKind.PERCENT }],

  [12_320, {}],
  [12_320, { numberKind: NumberKind.DOLLAR }],
  [12_320, { numberKind: NumberKind.PERCENT }],
  [12.23, {}],
  [12.23, { numberKind: NumberKind.DOLLAR }],
  [12.23, { numberKind: NumberKind.PERCENT }],

  [0.001432, {}],
  [0.001423, { numberKind: NumberKind.DOLLAR }],
  [0.001423, { numberKind: NumberKind.PERCENT }],

  [0.000_000_234_32, {}],
  [0.000_000_234_32, { numberKind: NumberKind.DOLLAR }],
  [0.000_000_234_32, { numberKind: NumberKind.PERCENT }],
];

describe("default formatter, throws on invalid inputs", () => {
  errorCases.forEach(([input, options]) => {
    it(`throws an errof for input: ${input}`, () => {
      const formatter = new IntTimesPowerOfTenFormatter([input], {
        ...baseOptions,
        ...options,
        ...{ onInvalidInput: "throw" },
      });
      expect(() => formatter.stringFormat(input)).toThrow();
    });
  });
});
