import {
  type FormatterOptionsCommon,
  type FormatterRangeSpecsStrategy,
  NumberKind,
  type RangeFormatSpec,
} from "../humanizer-types";

const bigNumberRangeSpec: RangeFormatSpec[] = [
  {
    minMag: -10,
    supMag: -4,
    maxDigitsRight: 2,
    baseMagnitude: 0,
    overrideValue: {
      int: "0",
      dot: ".",
      frac: "0001",
      prefix: "<",
      suffix: "",
    },
  },
  {
    minMag: -4,
    supMag: 0,
    maxDigitsLeft: 0,
    maxDigitsRight: 4,
    baseMagnitude: 0,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 0,
    supMag: 3,
    maxDigitsLeft: 3,
    maxDigitsRight: 2,
    baseMagnitude: 0,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 3,
    supMag: 5,
    maxDigitsRight: 0,
    useTrailingDot: false,
    baseMagnitude: 0,
    maxDigitsLeft: 5,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 5,
    supMag: 6,
    maxDigitsRight: 0,
    useTrailingDot: false,
    baseMagnitude: 3,
    maxDigitsLeft: 3,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 6,
    supMag: 7,
    maxDigitsRight: 2,
    baseMagnitude: 6,
    maxDigitsLeft: 1,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 7,
    supMag: 8,
    maxDigitsRight: 1,
    baseMagnitude: 6,
    maxDigitsLeft: 2,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 8,
    supMag: 9,
    maxDigitsRight: 0,
    baseMagnitude: 6,
    useTrailingDot: false,
    maxDigitsLeft: 3,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 9,
    supMag: 10,
    maxDigitsRight: 2,
    baseMagnitude: 9,
    maxDigitsLeft: 1,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 10,
    supMag: 11,
    maxDigitsRight: 1,
    baseMagnitude: 9,
    maxDigitsLeft: 2,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 11,
    supMag: 12,
    maxDigitsRight: 0,
    baseMagnitude: 9,
    useTrailingDot: false,
    maxDigitsLeft: 3,
    padWithInsignificantZeros: false,
  },
];

export const bigNumDefaultFormattingOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  numberKind: NumberKind.ANY,
  rangeSpecs: bigNumberRangeSpec,
  defaultMaxDigitsRight: 2,
  upperCaseEForExponent: true,
};

export const bigNumPercentOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  rangeSpecs: bigNumberRangeSpec,
  defaultMaxDigitsRight: 2,
  upperCaseEForExponent: true,
  numberKind: NumberKind.PERCENT,
};

export const bigNumCurrencyOptions = (
  numberKind: NumberKind,
): FormatterOptionsCommon & FormatterRangeSpecsStrategy => ({
  rangeSpecs: bigNumberRangeSpec,
  defaultMaxDigitsRight: 2,
  upperCaseEForExponent: true,
  numberKind,
});
