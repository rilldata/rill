export type FormatterOptionsV1 = {
  magnitudeStrategy:
    | "unlimited"
    | "unlimitedDigitTarget"
    | "largest"
    | "largestWithDigitTarget"
    | "defaultStrategy"
    | "defaultStrategyProposal-2023-03-02";
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
  useMaxDigitsRightIfSuffix: boolean;
  maxDigitsRightIfSuffix: number;

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
