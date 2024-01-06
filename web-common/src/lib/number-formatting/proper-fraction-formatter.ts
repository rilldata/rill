import { NumberKind, type NumberParts } from "./humanizer-types";
import { formatMeasurePercentageDifference } from "./percentage-formatter";
import { PerRangeFormatter } from "./strategies/per-range";

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
      `formatProperFractionAsPercent received invalid input: ${value}. Value must be between 0 and 1.`,
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
