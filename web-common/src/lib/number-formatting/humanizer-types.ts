import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

/**
 * This enum represents all of the valid strings that can be
 * used in the `format_preset` field of a measure definition.
 */
export enum FormatPreset {
  /**
   * In absence of a format preset, none is applied.
   */
  NONE = "none",
  HUMANIZE = "humanize",
  CURRENCY_USD = "currency_usd",
  CURRENCY_EUR = "currency_eur",
  PERCENTAGE = "percentage",
  INTERVAL = "interval_ms",
}

/**
 * This enum represents the semantic kind of the number being
 * handled (which is not the same thing as how the number is
 * formatted, though it can inform formatting).
 *
 * NOTE: (brendan, Jan 2024)
 * Requirements have changed since this was written,
 * and it's due for a rethinking. Based on experience and
 * requirements surfaced over the past year, recommended
 * approach would be to replace NumberKind with
 * something that makes the following concepts orthogonal
 * - units (dollar, euro, percent, etc)
 * - underlying conceptual number type (integer, real, 2-digit decimal).
 *   Note that in JS, these are all _stored_ as floats, but retaining
 *   the conceptual type is important for presentation
 * - formatting precision -- this can vary based on context,
 *   and could include options like
 *     - full number (which might still involve rounding off floating
 *       point errors for a Decimal)
 *     - some number of significant digits with letter suffixes
 *     - single digit time power of ten representation
 *     - etc...
 */
export enum NumberKind {
  /**
   * A real number with units of US Dollars. Note that this
   * does not imply any restriction on the range of the number;
   * ANY positive or negative real number of ANY SIZE can have
   * this units.
   */
  DOLLAR = "DOLLAR",

  /**
   * A real number with units of EUROS. Note that this
   * does not imply any restriction on the range of the number;
   * ANY positive or negative real number of ANY SIZE can have
   * this units.
   */
  EURO = "EURO",

  /**
   * A real number with units of "%". Note that this
   * does not imply any restriction on the range of the number;
   * ANY positive or negative real number of ANY SIZE can have
   * these units.
   * Additionally, `PERCENT` NumberKind assumes numbers have not
   * already been multiplied by 100; this will need to be applied
   * for formatting.
   */
  PERCENT = "PERCENT",

  /**
   * A real number that represents a time interval with
   * millisecond units.
   * This is a special case that is handled
   * by a custom formatter.
   */
  INTERVAL = "INTERVAL",

  /**
   * A generic real number that can be formatted in any way.
   */
  ANY = "ANY",
}

/**
 * This function converts a FormatPreset to a NumberKind.
 */
export const formatPresetToNumberKind = (type: FormatPreset) => {
  switch (type) {
    case FormatPreset.CURRENCY_USD:
      return NumberKind.DOLLAR;

    case FormatPreset.CURRENCY_EUR:
      return NumberKind.EURO;

    case FormatPreset.PERCENTAGE:
      return NumberKind.PERCENT;

    case FormatPreset.INTERVAL:
      return NumberKind.INTERVAL;

    case FormatPreset.NONE:
    case FormatPreset.HUMANIZE:
      return NumberKind.ANY;
    default:
      console.warn(
        `All FormatPreset variants must be explicity handled in formatPresetToNumberKind, got ${
          type === "" ? "empty string" : type
        }`,
      );
      return NumberKind.ANY;
  }
};

/**
 * Gets the NumberKind for a measure, based on its formatPreset.
 *
 * This wrapper around formatPresetToNumberKind allows that innner
 * function to maintain a more strict type signature.
 */
export const numberKindForMeasure = (measure: MetricsViewSpecMeasure) => {
  if (
    !measure ||
    measure.formatPreset === undefined ||
    measure.formatPreset === ""
  ) {
    // If no preset is specified, default to ANY
    return NumberKind.ANY;
  }
  return formatPresetToNumberKind(measure.formatPreset as FormatPreset);
};

export type LocaleConfig = {
  decimal?: string;
  thousands?: string;
  grouping?: number[];
};

export type NumberParts = {
  neg?: "-";
  currencySymbol?: "$" | "€" | string;
  int: string;
  dot: "" | ".";
  frac: string;
  suffix: string;
  prefix?: string;
  percent?: "%";
  approxZero?: boolean;
  locale?: LocaleConfig;
};

/**
 * Options common to all formatting strategies
 */
export type FormatterOptionsCommon = {
  /**
   * The kind of number being formatted
   */

  numberKind: NumberKind;

  /**
   * If true, pad numbers with insignificant zeros in order
   * to have a consistent number of digits to the right of the
   * decimal point
   */
  padWithInsignificantZeros?: boolean;

  /**
   * If `true`, use upper case "E" for exponential notation.
   * If `false` or `undefined`, use lower case "e"
   */
  upperCaseEForExponent?: boolean;
};

/**
 * This strategy does not apply formatting to the _number itself_,
 * but does add units if needed.
 */
export type FormatterOptionsNoneStrategy = FormatterOptionsCommon;

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
export type FormatterOptionsIntTimesPowerOfTenStrategy =
  FormatterOptionsCommon & {
    onInvalidInput?: "doNothing" | "throw" | "consoleWarn";
  };

/**
 * Specifies a set of formatting options for numbers within
 * a given order of magnitude range.
 */
export type RangeFormatSpec = {
  /**
   * Minimum order of magnitude for this range.
   * Target number must have OoM >= minMag.
   */
  minMag: number;

  /**
   * Supremum order of magnitude for this range.
   * Target number must have OoM OoM < supMag.
   */
  supMag: number;

  /**
   *Max number of digits left of decimal point.
   * If undefined, default is 3 digits
   */
  maxDigitsLeft?: number;

  /**
   * Max number of digits right of decimal point.
   */
  maxDigitsRight: number;

  /**
   * If set, this will be used as the order of magnitude
   * for formatting numbers in this range.
   * For example, if baseMagnitude=3, then we'd have:
   * - 1,000,000 => 1,000k
   * - 100 => .1k
   * If this is set to 0, numbers in this range
   * will be rendered as plain numbers (no suffix).
   * If not set, the engineering magnitude of `min` is used by default.
   */
  baseMagnitude?: number;

  /**
   * Whether or not to pad numbers with insignificant zeros. If undefined, treated as true
   */
  padWithInsignificantZeros?: boolean;

  /**
   * For a range with `maxDigitsRight=0`, by default a trailling
   * "." will be added if formatting causes some of a number's
   * true precision to be lost. For example, `123.234` with
   * `baseMagnitude=0` and `maxDigitsRight=0` will be rendered as
   * "123.", with the trailing "." retained to indicate that there
   * is additional precision that is not shown.
   *
   * If this is not desired, then setting `useTrailingDot=false` will
   * remove this decimal point--e.g., in the example above, `123.234`
   * will be formatted as just "123", with no decimal point.
   */
  useTrailingDot?: boolean;

  /**
   * If set, all numbers within the range use the number parts provided
   * ignoring all other spec instructions
   */
  overrideValue?: NumberParts;
};

/**
 * Strategy for formatting numbers based on order of magnitude ranges.
 *
 * `rangeSpecs` is a series of RangeFormatSpecs. Ranges may not overlap,
 * and there can be no gaps in coverage between the ranges that
 * are defined, though it is not required the the entire
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
export type FormatterRangeSpecsStrategy = FormatterOptionsCommon & {
  rangeSpecs: RangeFormatSpec[];
  defaultMaxDigitsRight: number;
};

export type FormatterFactoryOptions =
  | FormatterOptionsNoneStrategy
  | FormatterRangeSpecsStrategy;

export type NumPartPxWidthLookupFn = (str: string, isNumStr: boolean) => number;

export type FormatterFactory = (options: FormatterFactoryOptions) => Formatter;

export interface Formatter {
  options: FormatterFactoryOptions;
  stringFormat(x: number): string;
  partsFormat(x: number): NumberParts;
}

export type FormatterContext =
  | "table"
  | "unabridged"
  | "big-number"
  | "axis"
  | "tooltip";

export type FormatterContextSurface = Exclude<FormatterContext, "unabridged">;

export type ContextOptions = {
  none: FormatterRangeSpecsStrategy;
  currencyUsd: FormatterRangeSpecsStrategy;
  currencyEur: FormatterRangeSpecsStrategy;
  percent: FormatterRangeSpecsStrategy;
  humanize: FormatterRangeSpecsStrategy;
};
