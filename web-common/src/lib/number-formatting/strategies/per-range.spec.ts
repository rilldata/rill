import { FormatterFactoryOptions, NumberKind } from "../humanizer-types";
import { PerRangeFormatter } from "./per-range";

const invalidRangeOptions1: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    { minMag: 3, supMag: 3, maxDigitsRight: 0 },
    { minMag: -3, supMag: 3, maxDigitsRight: 3 },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

const invalidRangeOptions2: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    { minMag: 3, supMag: 2, maxDigitsRight: 0 },
    { minMag: -3, supMag: 3, maxDigitsRight: 3 },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

describe("range formatter constructor, throws if given an invalid range", () => {
  it(`should throw`, () => {
    expect(
      () => new PerRangeFormatter([100.12], invalidRangeOptions1)
    ).toThrow();
  });
  it(`should throw`, () => {
    expect(
      () => new PerRangeFormatter([100.12], invalidRangeOptions2)
    ).toThrow();
  });
});

const overlappingRangeOptions1: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    { minMag: 2, supMag: 6, maxDigitsRight: 0 },
    { minMag: -3, supMag: 3, maxDigitsRight: 3 },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

const overlappingRangeOptions2: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    { minMag: 2, supMag: 6, maxDigitsRight: 0 },
    { minMag: -3, supMag: 3, maxDigitsRight: 3 },
    { minMag: -6, supMag: -3, maxDigitsRight: 0 },
    { minMag: 6, supMag: 10, maxDigitsRight: 0 },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

describe("range formatter constructor, throws if given overlapping ranges", () => {
  it(`should throw`, () => {
    expect(
      () => new PerRangeFormatter([100.12], overlappingRangeOptions1)
    ).toThrow();
  });
  it(`should throw`, () => {
    expect(
      () => new PerRangeFormatter([100.12], overlappingRangeOptions2)
    ).toThrow();
  });
});

const gappedRangeOptions1: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    { minMag: 6, supMag: 9, maxDigitsRight: 0 },
    { minMag: -3, supMag: 3, maxDigitsRight: 3 },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

describe("range formatter constructor, throws if given gap in range coverage", () => {
  it(`should throw`, () => {
    expect(
      () => new PerRangeFormatter([100.12], gappedRangeOptions1)
    ).toThrow();
  });
});

const mar2ProposalOptions: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    {
      minMag: 3,
      supMag: 6,
      maxDigitsLeft: 6,
      maxDigitsRight: 0,
      padWithInsignificantZeros: true,
      baseMagnitude: 0,
    },
    {
      minMag: -3,
      supMag: 3,
      maxDigitsRight: 3,
      baseMagnitude: 0,
      padWithInsignificantZeros: true,
    },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

const mar2ProposalTestCases: [number, string][] = [
  // integers
  [999_999_999, "1.00B"],
  [12_345_789, "12.35M"],
  [2_345_789, "2.35M"],
  [999_999, "999,999"],
  [345_789, "345,789"],
  [45_789, "45,789"],
  [5_789, "5,789"],
  [999, "999.000"],
  [789, "789.000"],
  [89, "89.000"],
  [9, "9.000"],
  [0, "0"],
  [-999_999_999, "-1.00B"],
  [-12_345_789, "-12.35M"],
  [-2_345_789, "-2.35M"],
  [-999_999, "-999,999"],
  [-345_789, "-345,789"],
  [-45_789, "-45,789"],
  [-5_789, "-5,789"],
  [-999, "-999.000"],
  [-789, "-789.000"],
  [-89, "-89.000"],
  [-9, "-9.000"],
  [-0, "0"],

  // non integers
  [999_999_999.1234686, "1.00B"],
  [12_345_789.1234686, "12.35M"],
  [2_345_789.1234686, "2.35M"],
  [999_999.1234686, "999,999."],
  [345_789.1234686, "345,789."],
  [45_789.1234686, "45,789."],
  [5_789.1234686, "5,789."],
  [999.1234686, "999.123"],
  [789.1234686, "789.123"],
  [89.1234686, "89.123"],
  [9.1234686, "9.123"],
  [0.1234686, "0.123"],
  [-999_999_999.1234686, "-1.00B"],
  [-12_345_789.1234686, "-12.35M"],
  [-2_345_789.1234686, "-2.35M"],
  [-999_999.1234686, "-999,999."],
  [-345_789.1234686, "-345,789."],
  [-45_789.1234686, "-45,789."],
  [-5_789.1234686, "-5,789."],
  [-999.1234686, "-999.123"],
  [-789.1234686, "-789.123"],
  [-89.1234686, "-89.123"],
  [-9.1234686, "-9.123"],
  [-0.1234686, "-0.123"],

  // infinitesimals

  [0.00095, "0.001"],
  [0.000999999, "0.001"],
  [0.00012335234, "123.35e-6"],
  [0.000_000_999999, "1.00e-6"],
  [0.000_000_02341253, "23.41e-9"],
  [0.000_000_000_999999, "1.00e-9"],

  // padding with insignificant zeros
  [9.1, "9.100"],
  [9.12, "9.120"],
];

describe("range formatter, using options for 2022-03-02 proposal `.stringFormat()`", () => {
  mar2ProposalTestCases.forEach(([input, output]) => {
    it(`returns the correct string in case: ${input}`, () => {
      const formatter = new PerRangeFormatter([input], mar2ProposalOptions);
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});

const mar2ProposalNoZeroPadOptions: FormatterFactoryOptions = {
  strategy: "perRange",
  rangeSpecs: [
    {
      minMag: 3,
      supMag: 6,
      maxDigitsLeft: 6,
      maxDigitsRight: 0,
      padWithInsignificantZeros: false,
      baseMagnitude: 0,
    },
    {
      minMag: -3,
      supMag: 3,
      maxDigitsRight: 3,
      baseMagnitude: 0,
      padWithInsignificantZeros: false,
    },
  ],
  defaultMaxDigitsRight: 2,
  numberKind: NumberKind.ANY,
};

const mar2ProposalNoZeroPadTestCases: [number, string][] = [
  // integers
  [999_999_999, "1.00B"],
  [12_345_789, "12.35M"],
  [2_345_789, "2.35M"],
  [999_999, "999,999"],
  [345_789, "345,789"],
  [45_789, "45,789"],
  [5_789, "5,789"],
  [999, "999"],
  [789, "789"],
  [89, "89"],
  [9, "9"],
  [0, "0"],
  [-999_999_999, "-1.00B"],
  [-12_345_789, "-12.35M"],
  [-2_345_789, "-2.35M"],
  [-999_999, "-999,999"],
  [-345_789, "-345,789"],
  [-45_789, "-45,789"],
  [-5_789, "-5,789"],
  [-999, "-999"],
  [-789, "-789"],
  [-89, "-89"],
  [-9, "-9"],
  [-0, "0"],
];

describe("range formatter, using options for 2022-03-02 proposal and NO padding with insignificant zeros `.stringFormat()`", () => {
  mar2ProposalNoZeroPadTestCases.forEach(([input, output]) => {
    it(`returns the correct string in case: ${input}`, () => {
      const formatter = new PerRangeFormatter(
        [input],
        mar2ProposalNoZeroPadOptions
      );
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});
