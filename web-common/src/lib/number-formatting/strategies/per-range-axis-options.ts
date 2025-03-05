import {
  type FormatterOptionsCommon,
  type FormatterRangeSpecsStrategy,
  NumberKind,
  type RangeFormatSpec,
} from "../humanizer-types";

const axisRangeSpec: RangeFormatSpec[] = [
  {
    minMag: -4,
    supMag: -2,
    maxDigitsRight: 3,
    baseMagnitude: 0,
    padWithInsignificantZeros: false,
  },
  {
    minMag: -2,
    supMag: 3,
    maxDigitsRight: 2,
    baseMagnitude: 0,
    useTrailingDot: false,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 3,
    supMag: 11,
    maxDigitsRight: 2,
    baseMagnitude: 0,
    useTrailingDot: false,
    padWithInsignificantZeros: false,
  },
];

export const axisDefaultFormattingOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  numberKind: NumberKind.ANY,
  rangeSpecs: axisRangeSpec,
  defaultMaxDigitsRight: 0,
  padWithInsignificantZeros: false,
};

export const axisPercentOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  rangeSpecs: axisRangeSpec,
  defaultMaxDigitsRight: 0,
  numberKind: NumberKind.PERCENT,
};

export const axisCurrencyOptions = (
  numberKind: NumberKind,
): FormatterOptionsCommon & FormatterRangeSpecsStrategy => ({
  rangeSpecs: axisRangeSpec,
  defaultMaxDigitsRight: 0,
  numberKind,
});
