export type FormatterOptionsV1 = {
  magnitudeStrategy:
    | "unlimited"
    | "unlimitedDigitTarget"
    | "largestWithDigitTarget";
  digitTarget: number;
  digitTargetPadWithInsignificantZeros: boolean;
  usePlainNumsForThousands: boolean;
  usePlainNumsForThousandsOneDecimal: boolean;
  usePlainNumForThousandths: boolean;
  usePlainNumForThousandthsPadZeros: boolean;
  truncateThousandths: boolean;
  truncateTinyOrdersIfBigOrderExists: boolean;
  zeroHandling: "exactZero" | "noSpecial" | "zeroDot";
  maxTotalDigits: number;
  maxDigitsLeft: number;
  maxDigitsRight: number;
  minDigitsNonzero: number;
  nonIntegerHandling: "none" | "oneDigit" | "trailingDot";
  formattingUnits: "none" | "$" | "%";
  specialDecimalHandling: "noSpecial" | "alwaysTwoDigits" | "neverOneDigit";

  alignDecimalPoints: boolean;
  alignSuffixes: boolean;
  suffixPadding: number;
  lowerCaseEForEng: boolean;
  showMagSuffixForZero: boolean;
};

export type FormatterOptionsV1Partial = Partial<FormatterOptionsV1>;
