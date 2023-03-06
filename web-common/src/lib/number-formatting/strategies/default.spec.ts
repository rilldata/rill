import {
  FormatterFactoryOptions,
  NumberKind,
  NumberParts,
} from "../humanizer-types";
import { DefaultHumanizer } from "./default";

type TestArgs = [number, number, number, boolean, boolean?];

const baseOptions: FormatterFactoryOptions = {
  strategy: "default",
  padWithInsignificantZeros: true,
  numberKind: NumberKind.ANY,
  maxDigitsRightSmallNums: 3,
  maxDigitsRightSuffixNums: 2,
};

const testCases: [
  number,
  {
    maxDigitsRightSmallNums?: number;
    maxDigitsRightSuffixNums?: number;
    padWithInsignificantZeros?: boolean;
    numberKind: NumberKind;
  },
  string
][] = [
  [12_345_789, , "12.35M"],
  [12_345_789, { numberKind: NumberKind.DOLLAR }, "$12.35M"],
  [12_345_789, { numberKind: NumberKind.PERCENT }, "1.23B%"],

  [12_345.789012, , "12346."],
  [12_345.789012, { numberKind: NumberKind.DOLLAR }, "$12346."],
  [12_345.789012, { numberKind: NumberKind.PERCENT }, "1.23M%"],

  [12.345789012, , "12.346"],
  [12.345789012, { numberKind: NumberKind.DOLLAR }, "$12.35"],
  [12.345789012, { numberKind: NumberKind.PERCENT }, "1235.%"],

  [0.0012345789012, , "0.001"],
  [0.0012345789012, { numberKind: NumberKind.DOLLAR }, "$1.23e-3"],
  [0.0012345789012, { numberKind: NumberKind.PERCENT }, "0.123%"],

  [0.000_000_2345789012, , "234.58e-9"],
  [0.000_000_2345789012, { numberKind: NumberKind.DOLLAR }, "$234.58e-9"],
  [0.000_000_2345789012, { numberKind: NumberKind.PERCENT }, "23.46e-6%"],
];

describe("default formatter, default options `.stringFormat()`", () => {
  testCases.forEach(([input, options, output]) => {
    it(`returns the correct split string in case: ${input}`, () => {
      const formatter = new DefaultHumanizer([input], {
        ...baseOptions,
        ...options,
      });
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});
