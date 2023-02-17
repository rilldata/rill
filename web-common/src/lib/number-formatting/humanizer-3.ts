export type NumberFormatter = (x: number) => RichFormatNumber;

export type NumberStringParts = {
  int: string;
  dot: "" | ".";
  frac: string;
  suffix: string;
};

export type NumericRange = {
  min: number;
  max: number;
};

export type FormatterPxWidths = {
  int: number;
  dot: number;
  frac: number;
  suffix: number;
};

export type RichFormatNumber = {
  number: number;
  splitStr: NumberStringParts;
};

export type FormatterFactoryOptions = (
  | {
      strategy: "largestMagnitude";
    }
  | {
      strategy: "multipleMagnitudes";
      maxDigitsLeft: number;
      maxDigitsRight: number;
      minDigitsNonzero: number;
    }
) & {
  // Options common to all strategies

  //
  maxTotalDigits: number;

  // (Not recommended)
  padWithInSignificantZeros: boolean;

  // (Not recommended)
  dropSignificantTrailingZeros: boolean;

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
  // method for formatting exact zeros
  zeroHandling: "noSpecial" | "trailingDot" | "zeroOnly";
  pxWidthLookup: undefined | FormatterPxWidths;
};

export type FormatterFactoryOutput = {
  formatter: NumberFormatter;
  options: FormatterFactoryOptions;

  // Maximum formatted widths in this sample.
  // Will be `undefined` if a pixel with lookup
  // function is not given in the options.
  maxWidthsInSample?: FormatterPxWidths;
  // Maximum possible formatted widths given the options used.
  // It is possible that no actual number in the sample
  // will be this wide, but this is the upper limit on the width
  // that can be reached given these formatting options. This44
  // may be useful if additional numbers will be added to a sample
  // later and as in the column would be undesirable.
  // Will be `undefined` if a pixel with lookup
  // function is not given in the options.
  maxWidthPossible?: FormatterPxWidths;

  // the largest order of magnitude of any number in this data set
  largestMagnitude: number;
  // the Order of magnitude of the most precise digit in any
  // number from this data set
  smallestMagnitude: number;
  // the min and max of this data set
  range: NumericRange;
};

export type NumPartPxWidthLookupFn = (str: string, isNumStr: boolean) => number;

export type FormatterFactory = (
  sample: number[],
  options
) => FormatterFactoryOutput;
