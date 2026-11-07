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
  // Scale each magnitude band to its own SI unit (k/M/B) and keep one
  // decimal, so tick labels stay compact without rounding away a meaningful
  // digit. A single [3, 11) band with baseMagnitude 0 could not do this: it
  // produced a 4+ digit integer that failed validation, fell through to the
  // 0-decimal default path, and rounded e.g. 1,500,000 up to "2M".
  {
    minMag: 3,
    supMag: 6,
    maxDigitsRight: 1,
    baseMagnitude: 3,
    maxDigitsLeft: 3,
    useTrailingDot: false,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 6,
    supMag: 9,
    maxDigitsRight: 1,
    baseMagnitude: 6,
    maxDigitsLeft: 3,
    useTrailingDot: false,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 9,
    supMag: 12,
    maxDigitsRight: 1,
    baseMagnitude: 9,
    maxDigitsLeft: 3,
    useTrailingDot: false,
    padWithInsignificantZeros: false,
  },
];

export const axisDefaultFormattingOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  numberKind: NumberKind.ANY,
  rangeSpecs: axisRangeSpec,
  defaultMaxDigitsRight: 0,
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
