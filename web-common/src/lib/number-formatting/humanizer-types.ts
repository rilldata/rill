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

// FIXME: we can add these types back in if we want to implement
// alignment. If we decide that we don't want to pursue that,
// we can remove this commented code
// export type PxWidthLookupFn = (str: string | undefined) => number;
// export type FormatterWidths = {
//   left: number;
//   dot: number;
//   frac: number;
//   suffix: number;
// };

export type RichFormatNumber = {
  number: number;
  splitStr: NumberParts;
};

/**
 * This enum represents the *semantic* kind of the number being
 * handled, which is orthogonal to how it is being formatted.
 *
 * (Note that this contrasts with NicelyFormattedTypes)
 */
export enum NumberKind {
  DOLLAR = "DOLLAR",
  PERCENT = "PERCENT",
  ANY = "ANY",
}

/**
 * This is a no-op strategy
 */
export type FormatterOptionsNoneStrategy = {
  strategy: "none";
};

/**
 * Strategy for handling numbers that are guaranteed to be an
 * integer multiple of a power of ten, such as the output of
 * d3 scale ticks.
 *
 * The number will be formatted
 * with a short scale suffix or an or engineering order
 * of magnitude (a multiple of three). If the magnitude
 * is 10^0, no suffix is used.
 *
 * A formatter using this strategy can be set to throw an error
 * or log a warning if a of a non integer multiple of a power
 * of ten given as an input.
 */
export type FormatterOptionsIntTimesPowerOfTenStrategy = {
  strategy: "intTimesPowerOfTen";
  onInvalidInput?: "doNothing" | "throw" | "consoleWarn";
};

/**
 * The "default" strategy actaully delegates to a set of
 * pre-defined FormatterRangeSpecsStrategies, one for
 * each of the three NumberKinds currently supported.
 */
export type FormatterOptionsDefaultStrategy = {
  strategy: "default";
};

/**
 * Specifies a set of formatting options
 */
export type RangeFormatSpec = {
  // minimum order of magnitude for this range.
  // Target number must have OoM >= minMag.
  minMag: number;
  // supremum number for this range.
  // Target number must have OoM < supMag.
  supMag: number;

  // max number of digits left of decimal point
  // if undefined, default is 3 digits
  maxDigitsLeft?: number;
  // max number of digits right of decimal point
  maxDigitsRight: number;
  // This sets the order of magnitude used to format numbers
  // in this range. For example, if baseMagnitude=3, then we'd have:
  // - 1,000,000 => 1,000k
  // - 100 => .1k
  // If this is set to 0, numbers in this range
  // will be rendered as plain numbers (no suffix).
  // If not set, the engineering magnitude of `min` is used by default.
  baseMagnitude?: number;
  // if not set, treated as true
  padWithInsignificantZeros?: boolean;
  // if not set, treated as true
  useTrailingDot?: boolean;
};

/**
 * Strategy for formatting numbers based on order of magnitude ranges.
 *
 * `rangeSpecs` is a series of RangeFormatSpecs. Ranges may not overlap,
 * and there can be no gaps in coverage between the ranges that
 * are defined, though the it is not required the the entire
 * number line be covered--defaults will be used outside of the
 * covered range.
 *
 * Each order of magnitude range must supply a minimum and supremum order
 * of magnitude that sets what numbers will be formatted using that range's
 * rules, and must also set a maximum number of RHS digits. Other formatting
 * rules may optionally be set as well, see RangeFormatSpec.
 *
 * It may be possible to define sets of rules that are incompatible if very
 * unusual parameter values have been supplied in RangeFormatSpec. The formatter
 * constructor will throw an errot in the following cases:
 * - If any range has minMag >= supMag
 * - if any ranges overlap
 * - if there are gaps between ranges
 *
 * Note that defaultMaxDigitsRight can be set by the user, but
 * it is not possible to set a maximum number of left hand digits,
 * because this can conflict with engineering-style order of magnitude
 * groupings if anything other than three is used. Therefore,
 * if more than three digits are desired left of the decimal point, an
 * explicit range must be set with maxDigitsLeft.
 */
export type FormatterRangeSpecsStrategy = {
  strategy: "perRange";
  rangeSpecs: RangeFormatSpec[];
  defaultMaxDigitsRight: number;
};

// FIXME: These strategies still need production grade implementation.
// If we decide not to implement these strategis for production,
// this code can be removed.
// export type FormatterOptionsLargestMag = {
//   // options specific to the largestMagnitude strategy
//   strategy: "largestMagnitude";
// };
// export type FormatterOptionsDigitBudget = {
//   // options specific to the multipleMagnitudes strategy
//   strategy: "digitBudget";
//   maxDigitsLeft: number;
//   maxDigitsRight: number;
//   minDigitsNonzero: number;

//   // Method for showing that non-integers have a fractional
//   // part if they would otherwise be rounded such that they
//   // have no fractional digits.
//   // "none": don't do anything special.
//   // Ex: 21379.23 with max 5 digits would be "21379"
//   // "trailingDot": add a trailing decimal point if a non-integer
//   // would be truncated to the e0 digit.
//   // Ex: 21379.23 with max 5 digits would be "21379."
//   // "reserveDigit": Always reserve one digit from the max digit
//   // budget to show a digit of precision after the decimal point.
//   // Ex: 21379.23 with max 5 digits would require an order of mag
//   // suffix, e.g. "21.379 k"; or with max 6 digits "21379.2"
//   nonIntHandling: "none" | "trailingDot" | "reserveDigit";
// };

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
  // FIXME: These strategies still need production grade implementation.
  // If we decide not to implement these strategis for production,
  // this code can be removed.
  // | FormatterOptionsDigitBudget
  // | FormatterOptionsLargestMag
  | FormatterOptionsIntTimesPowerOfTenStrategy
  | FormatterRangeSpecsStrategy
  | FormatterOptionsDefaultStrategy
) &
  FormatterOptionsCommon;

export type NumPartPxWidthLookupFn = (str: string, isNumStr: boolean) => number;

export type FormatterFactory = (
  sample: number[],
  options: FormatterFactoryOptions
) => Formatter;

export interface Formatter {
  options: FormatterFactoryOptions;

  stringFormat(x: number): string;

  partsFormat(x: number): NumberParts;

  // FIXME: we can add these parts of the interface back in if we want to implement
  // alignment. If we decide that we don't want to pursue that,
  // we can remove this commented code
  // largestPossibleNumberStringParts: NumberParts;
  // maxPxWidthsSampled(): FormatterWidths;
  // maxPxWidthsPossible(): FormatterWidths;
  // maxCharWidthsSampled(): FormatterWidths;
  // maxCharWidthsPossible(): FormatterWidths;
}
