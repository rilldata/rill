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
  DOLLAR = "DOLLAR",
  PERCENT = "PERCENT",
  ANY = "ANY",
}

export type FormatterOptionsNoneStrategy = {
  strategy: "none";
};

/**
 * Strategy for handling numbers that are guaranteed to be an
 * integer multiple of a power of ten, such as the output of
 * d3 scale ticks.
 *
 * The number will be formatted
 * with short scale a short scale suffix or an or engineering order
 * of magnitude (floored multiples of three). If the floored magnitude
 * is 10^0, no suffix is used.
 *
 * In case of a non integer multiple of a power of ten being passed
 * into this strategy, an error or warning may be shown.
 */
export type FormatterOptionsIntTimesPowerOfTenStrategy = {
  strategy: "intTimesPowerOfTen";
  onInvalidInput?: "doNothing" | "throw" | "consoleWarn";
};

export type FormatterOptionsDefaultStrategy = {
  strategy: "default";
};

export type RangeFormatSpec = {
  /**
   * this set of parameters check whether the number should
   * have this formatting applied.
   *
   * Note that we use orders of magnitude here rather than
   * just numbers because that clarifies the handling of
   * negative and positive numbers without repetition.
   */
  // minimum number for this range.
  // Target number must be >= min.
  minMag: number;
  // supremum number for this range.
  // Target number must be < sup.
  supMag: number;

  /**
   * this set of parameters controls the formatting used
   * for numbers in this range
   */

  // max number of digits left of decimal point
  // if umdefined, default is 3 digits
  maxDigitsLeft?: number;
  // max number of digits right of decimal point
  maxDigitsRight: number;
  // This sets the order of magnitude used to format numbers
  // in this range. If this is set to 0, numbers in this range
  // will be rendered as plain numbers (no suffix).
  // If not set, the engineering magnitude of `min` is used by default.
  baseMagnitude?: number;
  // if not set, treated as true
  padWithInsignificantZeros?: boolean;
  // if not set, treated as true
  useTrailingDot?: boolean;
};

/**
 * Note that defaultMaxDigitsRight can be set by the user, but
 * it is not possible to set a maximum number of left hand digits,
 * because this can conflict with engineering-style order of magnitude
 * groupings if anything other than three is used. Therefore,
 * if more than three digits are desired left of the decimal point, an
 * explicit range must be set.
 */
export type FormatterOptionsPerRangeStrategy = {
  strategy: "perRange";
  /**
   * This is a series of RangeFormatSpecs. Ranges may not overlap,
   * and there can be nogaps in coverage between the ranges that
   * are defined, though the it is not required the the entire
   * number line be covered--defaults will be used outside of the
   * covered range.
   */
  rangeSpecs: RangeFormatSpec[];
  defaultMaxDigitsRight: number;
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
  padWithInsignificantZeros?: boolean;

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

  // If `true`, use commas to group thousands when applicable;
  // If `false` or `undefined`, no commas.
  useCommas?: boolean;
};

export type FormatterFactoryOptions = (
  | FormatterOptionsNoneStrategy
  | FormatterOptionsIntTimesPowerOfTenStrategy
  | FormatterOptionsPerRangeStrategy
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
