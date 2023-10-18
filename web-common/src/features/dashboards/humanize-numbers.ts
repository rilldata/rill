// Create types and then present an appropriate string
// Current dash persion has `prefix` key in JSON to add currecny etc.
// We can provide a dropdown option in the table?? or regex??

import {
  NumberKind,
  NumberParts,
} from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { PerRangeFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/per-range";

/**
 * This enum represents all of the valid strings that can be
 * used in the `format_preset` field of a measure definition.
 */
export enum FormatPreset {
  HUMANIZE = "humanize",
  NONE = "none",
  CURRENCY = "currency_usd",
  PERCENTAGE = "percentage",
  INTERVAL = "interval_ms",
}

// NOTE: the following are adapters that I think fit the API
// used by the existing humanizer, but I'm not sure of the
// exact details, nor am I totally confident about the options
// passed in at all the relevant call sites, so I've added
// thes adapters rather than just pave over the existing functions.
// This really needs to be reviewed by Dhiraj, at which point we
// can deprecate any left over code that is no longer needed.

export const formatPresetToNumberKind = (type: FormatPreset | string) => {
  switch (type) {
    case FormatPreset.CURRENCY:
      return NumberKind.DOLLAR;

    case FormatPreset.PERCENTAGE:
      return NumberKind.PERCENT;

    case FormatPreset.INTERVAL:
      return NumberKind.INTERVAL;

    default:
      // captures:
      // FormatPreset.NONE
      // FormatPreset.HUMANIZE
      return NumberKind.ANY;
  }
};

/**
 * Formatter for the comparison percentage differences.
 * Input values are given as proportions, not percentages
 * (not yet multiplied by 100). However, inputs may be
 * any real number, not just a proper fraction, so negative
 * values and values of arbitrarily large magnitudes must be
 * supported.
 */
export function formatMeasurePercentageDifference(value: number): NumberParts {
  if (value === 0) {
    return { percent: "%", int: "0", dot: "", frac: "", suffix: "" };
  } else if (value < 0.005 && value > 0) {
    return {
      percent: "%",
      int: "0",
      dot: "",
      frac: "",
      suffix: "",
      approxZero: true,
    };
  } else if (value > -0.005 && value < 0) {
    return {
      percent: "%",
      neg: "-",
      int: "0",
      dot: "",
      frac: "",
      suffix: "",
      approxZero: true,
    };
  }

  const factory = new PerRangeFormatter([], {
    strategy: "perRange",
    rangeSpecs: [
      {
        minMag: -2,
        supMag: 3,
        maxDigitsRight: 1,
        baseMagnitude: 0,
        padWithInsignificantZeros: false,
      },
    ],
    defaultMaxDigitsRight: 0,
    numberKind: NumberKind.PERCENT,
  });

  return factory["partsFormat"](value);
}

/**
 * This function is used to format proper fractions, which
 * must be between 0 and 1, as percentages. It is used in
 * formatting the percentage of total column, as well as
 * other contexts where the input number is guaranteed to
 * be a proper fraction.
 *
 * If the input number is not a proper fraction, this function
 * will `console.warn` (since this is not worth crashing over)
 * and use formatMeasurePercentageDifference
 * instead, though that will likely result in a badly formatted
 * output, since formatting of proper fractions may make
 * assumptions that are violated by improper fractions.
 */
export function formatProperFractionAsPercent(value: number): NumberParts {
  if (value < 0 || value > 1) {
    console.warn(
      `formatProperFractionAsPercent received invalid input: ${value}. Value must be between 0 and 1.`
    );
    return formatMeasurePercentageDifference(value);
  }

  if (value < 0.01 && value !== 0) {
    return { percent: "%", int: "<1", dot: "", frac: "", suffix: "" };
  } else if (value === 0) {
    return { percent: "%", int: "0", dot: "", frac: "", suffix: "" };
  }
  const factory = new PerRangeFormatter([], {
    strategy: "perRange",
    rangeSpecs: [
      {
        minMag: -2,
        supMag: 3,
        maxDigitsRight: 1,
        baseMagnitude: 0,
        padWithInsignificantZeros: false,
      },
    ],
    defaultMaxDigitsRight: 0,
    numberKind: NumberKind.PERCENT,
  });

  return factory["partsFormat"](value);
}
