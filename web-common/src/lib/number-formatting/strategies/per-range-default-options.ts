import {
  FormatterOptionsCommon,
  FormatterOptionsPerRangeStrategy,
  NumberKind,
} from "../humanizer-types";

export const defaultGenericNumOptions: FormatterOptionsCommon &
  FormatterOptionsPerRangeStrategy = {
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
  FormatterOptionsPerRangeStrategy = {
  ...defaultGenericNumOptions,
  numberKind: NumberKind.PERCENT,
};

export const defaultDollarOptions: FormatterOptionsCommon &
  FormatterOptionsPerRangeStrategy = {
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
