import {
  type FormatterOptionsCommon,
  type FormatterRangeSpecsStrategy,
  NumberKind,
  type RangeFormatSpec,
} from "../humanizer-types";

const bigNumberRangeSpec: RangeFormatSpec[] = [
  {
    minMag: -6,
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
    maxDigitsRight: 4,
    baseMagnitude: 0,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 0,
    supMag: 3,
    maxDigitsRight: 2,
    baseMagnitude: 0,
    padWithInsignificantZeros: false,
  },
  {
    minMag: 3,
    supMag: 5,
    maxDigitsRight: 0,
    baseMagnitude: 0,
    maxDigitsLeft: 5,
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
