import { NumberKind, NumberParts } from "./humanizer-types";
import { PerRangeFormatter } from "./strategies/per-range";

/**
 * Formatter for the comparison percentage differences.
 * Input values are given as proportions, not percentages
 * (not yet multiplied by 100). However, inputs may be
 * any real number, not just a proper fraction, so negative
 * values and values of arbitrarily large magnitudes must be
 * supported.
 */
export function formatMeasurePercentageDifference(value: number): NumberParts {
  // no-op comment change to prove working
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

  const factory = new PerRangeFormatter({
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
