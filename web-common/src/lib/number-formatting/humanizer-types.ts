export type NumberFormatter = (x: number) => RichFormatNumber;

export type NumberParts = {
  neg?: "-";
  dollar?: "$";
  int: string;
  dot: "" | ".";
  frac: string;
  suffix: string;
  percent?: "%";
};

export type NumericRange = {
  min: number;
  max: number;
};

export type PxWidthLookupFn = (str: string | undefined) => number;

export type FormatterWidths = {
  left: number;
  dot: number;
  frac: number;
  suffix: number;
};

export type RichFormatNumber = {
  number: number;
  splitStr: NumberParts;
};

export enum NumberKind {
  DOLLAR,
  PERCENT,
  ANY,
}

export type FormatterOptionsNoneStrategy = {
  strategy: "none";
};

export type FormatterOptionsDefaultStrategy = {
  strategy: "default";
  // number of RHS digits for x s.t.: 1e-3 <= x < 1e6
  maxDigitsRightSmallNums: number;
  // number of RHS digits for numbers rendered with a suffix
  maxDigitsRightSuffixNums: number;
};

export type FormatterOptionsLargestMag = {
  // options specific to the largestMagnitude strategy
  strategy: "largestMagnitude";
};
export type FormatterOptionsDigitBudget = {
  // options specific to the multipleMagnitudes strategy
  strategy: "digitBudget";
  maxDigitsLeft: number;
  maxDigitsRight: number;
  minDigitsNonzero: number;

  // Method for showing that non-integers have a fractional
  // part if they would otherwise be rounded such that they
  // have no fractional digits.
  // "none": don't do anything special.
  // Ex: 21379.23 with max 5 digits would be "21379"
  // "trailingDot": add a trailing decimal point if a non-integer
  // would be truncated to the e0 digit.
  // Ex: 21379.23 with max 5 digits would be "21379."
  // "reserveDigit": Always reserve one digit from the max digit
  // budget to show a digit of precision after the decimal point.
  // Ex: 21379.23 with max 5 digits would require an order of mag
  // suffix, e.g. "21.379 k"; or with max 6 digits "21379.2"
  nonIntHandling: "none" | "trailingDot" | "reserveDigit";
};

export type FormatterOptionsCommon = {
  // Options common to all strategies

  // max number of digits to be shown for formatted numbers
  // maxTotalDigits: number;

  // The kind of number being formatted
  numberKind: NumberKind;

  // If true, pad numbers with insignificant zeros in order
  // to have a consistent number of digits to the right of the
  // decimal point
  padWithInsignificantZeros: boolean;

  // method for formatting exact zeros
  // "none": don't do anything special.
  // Ex: If the general option padWithInsignificantZeros is used such
  // that e.g. a 0 is rendered as "0.000", then if
  // this option is "none", the trailing zeros will be retained
  // "trailingDot": add a trailing decimal point to exact zeros "0."
  // "zeroOnly": render exact zeros as "0"
  // zeroHandling: "none" | "trailingDot" | "zeroOnly";

  // pxWidthLookupFn?: PxWidthLookupFn;

  // not actually used for formatting, but needed to calculate the
  // px sizes of maxWidthsInSample and maxWidthsPossible
  // alignDecimal?: boolean;

  // If `true`, use upper case "E" for exponential notation;
  // If `false` or `undefined`, use lower case
  upperCaseEForExponent?: boolean;
};

export type FormatterFactoryOptions = (
  | FormatterOptionsNoneStrategy
  | FormatterOptionsDefaultStrategy
  | FormatterOptionsDigitBudget
  | FormatterOptionsLargestMag
) &
  FormatterOptionsCommon;

// FIXME? --maybe deprecated

// export type FormatterFactoryOutput = {
//   formatter: NumberFormatter;
//   options: FormatterFactoryOptions;

//   // Maximum formatted pixel widths in this sample.
//   // Will be `undefined` if a pxWidthLookupFn
//   // is not given in the options.
//   maxPxWidthsInSample?: FormatterWidths;
//   // Maximum possible formatted widths given the options used.
//   // It is possible that no actual number in the sample
//   // will be this wide, but this is the upper limit on the width
//   // that can be reached given these formatting options. This
//   // may be useful if additional numbers will be added to a sample
//   // later and  resizing the column would be undesirable.
//   // Will be `undefined` if a pxWidthLookupFn
//   // is not given in the options.
//   maxPxWidthPossible?: FormatterWidths;

//   // Maximum formatted character widths in this sample.
//   maxCharWidthsInSample: FormatterWidths;
//   // Maximum possible formatted character widths given the options used.
//   // It is possible that no actual number in the sample
//   // will be this wide, but this is the upper limit on the width
//   // that can be reached given these formatting options. This
//   // may be useful if additional numbers will be added to a sample
//   // later and resizing the column would be undesirable.
//   maxCharWidthPossible: FormatterWidths;

//   // the largest order of magnitude of any number in this data set
//   largestMagnitude: number;
//   // the Order of magnitude of the most precise digit in any
//   // number from this data set
//   mostPreciseDigitMagnitude: number;
//   // the min and max of this data set
//   range: NumericRange;
// };

export type NumPartPxWidthLookupFn = (str: string, isNumStr: boolean) => number;

export type FormatterFactory = (
  sample: number[],
  options: FormatterFactoryOptions
) => Formatter;

export interface Formatter {
  options: FormatterFactoryOptions;
  largestPossibleNumberStringParts: NumberParts;

  stringFormat(x: number): string;

  partsFormat(x: number): NumberParts;

  // FIXME? -- will be needed if we want alignment
  // maxPxWidthsSampled(): FormatterWidths;
  // maxPxWidthsPossible(): FormatterWidths;

  // maxCharWidthsSampled(): FormatterWidths;
  // maxCharWidthsPossible(): FormatterWidths;
}
