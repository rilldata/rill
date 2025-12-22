import {
  type FormatterOptionsCommon,
  type FormatterRangeSpecsStrategy,
  NumberKind,
} from "../humanizer-types";

export const defaultNoFormattingOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  numberKind: NumberKind.ANY,
  rangeSpecs: [
    {
      minMag: -5,
      supMag: -2,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      overrideValue: {
        int: "",
        dot: ".",
        frac: "00",
        prefix: "~",
        suffix: "",
      },
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
      maxDigitsRight: 0,
      maxDigitsLeft: 12,
      baseMagnitude: 0,
      useTrailingDot: false,
      padWithInsignificantZeros: false,
    },
  ],
  defaultMaxDigitsRight: 2,
  upperCaseEForExponent: true,
};

export const defaultGenericNumOptions: FormatterOptionsCommon &
  FormatterRangeSpecsStrategy = {
  rangeSpecs: [
    {
      minMag: -5,
      supMag: -2,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      overrideValue: {
        int: "",
        dot: ".",
        frac: "00",
        prefix: "~",
        suffix: "",
      },
    },
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
  rangeSpecs: [
    {
      minMag: -6,
      supMag: -2,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      overrideValue: {
        int: "",
        dot: ".",
        frac: "00",
        prefix: "~",
        suffix: "",
      },
    },
    {
      minMag: -2,
      supMag: -1,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      padWithInsignificantZeros: false,
    },
    {
      minMag: -1,
      supMag: 3,
      maxDigitsRight: 1,
      baseMagnitude: 0,
      padWithInsignificantZeros: false,
    },
  ],
  defaultMaxDigitsRight: 1,
  numberKind: NumberKind.PERCENT,
};

export const defaultCurrencyOptions = (
  numberKind: NumberKind,
): FormatterOptionsCommon & FormatterRangeSpecsStrategy => ({
  rangeSpecs: [
    {
      minMag: -5,
      supMag: -2,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      overrideValue: {
        int: "",
        dot: ".",
        frac: "00",
        prefix: "~",
        suffix: "",
      },
    },
    {
      minMag: -2,
      supMag: 3,
      maxDigitsRight: 2,
      baseMagnitude: 0,
      padWithInsignificantZeros: true,
    },
  ],
  defaultMaxDigitsRight: 1,
  numberKind,
});
