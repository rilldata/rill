import {
  FormatterOptionsCommon,
  FormatterRangeSpecsStrategy,
  NumberKind,
} from "../humanizer-types";

export const defaultGenericNumOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  strategy: "perRange",
  rangeSpecs: [
    {
      minMag: -2,
      supMag: 3,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      padWithInsignificantZeros: false,
    },
  ],
  defaultMaxDigitsRight: 1,
  numberKind: NumberKind.ANY,
};

export const defaultPercentOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  strategy: "perRange",
  rangeSpecs: [
    {
      minMag: -2,
      supMag: 3,
      maxDigitsRight: 1,
      baseMagnitude: 0,
      padWithInsignificantZeros: false,
    },
  ],
  defaultMaxDigitsRight: 1,
  numberKind: NumberKind.PERCENT,
};

export const defaultDollarOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  strategy: "perRange",
  rangeSpecs: [
    {
      minMag: -2,
      supMag: 3,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      padWithInsignificantZeros: true,
    },
  ],
  defaultMaxDigitsRight: 1,
  numberKind: NumberKind.DOLLAR,
};
