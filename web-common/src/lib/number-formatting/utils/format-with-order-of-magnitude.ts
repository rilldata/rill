import type { NumberParts } from "../humanizer-types";
import { smallestPrecisionMagnitude } from "./smallest-precision-magnitude";

export const orderOfMagnitude = (x) => {
  if (x === 0) return 0;
  return Math.floor(Math.log10(Math.abs(x)));
};

export const orderOfMagnitudeEng = (x) => {
  return Math.round(Math.floor(orderOfMagnitude(x) / 3) * 3);
};

export const formatNumWithOrderOfMag = (
  x: number,
  newOrder: number,
  fractionDigits: number,
  // Set to true to pad with insignificant zeros.
  // Integers will be padded with zeros if this is set.
  padInsignificantZeros = false,
  // Set to `true` to leave a trailing "." in the case
  // of non-integers formatted to e0 with 0 fraction digits.
  // Even if this is `true` integers WILL NOT be formatted with a trailing "."
  trailingDot = false
): NumberParts => {
  if (typeof x !== "number") throw new Error("input must be a number");

  if (x === Infinity) return { int: "∞", dot: "", frac: "", suffix: "" };
  if (x === -Infinity) return { int: "-∞", dot: "", frac: "", suffix: "" };
  if (Number.isNaN(x)) return { int: "NaN", dot: "", frac: "", suffix: "" };

  const suffix = "E" + newOrder;
  let dot: "" | "." = ".";

  if (x === 0)
    return {
      int: "0",
      dot,
      frac: padInsignificantZeros ? "0".repeat(fractionDigits) : "",
      suffix,
    };

  if (padInsignificantZeros === false) {
    // get the OoM of the smallest digit
    const spm = smallestPrecisionMagnitude(x);
    // calculate the order of mag of the smallest precision digit
    // when represented in this new magnitude
    const newSpm = spm - newOrder;
    // Use the new order to set the final value of fractionDigits
    // to be used in the formatter.
    // Note that the the number of fractional digits to keep will be positive
    // only if the smallest precision OoM is negative.
    // Finally, if not using zero padding, we want to use the smaller of
    // (a) the number of digits needed for this smallest precision, or
    // (b) the target number of fraction digits

    // if padding would *add* zeros that are not needed,

    if (newSpm < 0) {
      fractionDigits = Math.min(-newSpm, fractionDigits);
    } else {
      fractionDigits = 0;
    }
  }

  const [int, frac] = Intl.NumberFormat("en-US", {
    maximumFractionDigits: fractionDigits,
    minimumFractionDigits: fractionDigits,
  })
    .format(x / 10 ** newOrder)
    //.replace(/,/g, "")
    .split(".");

  const nonInt = !Number.isInteger(x);

  dot =
    frac !== undefined || (fractionDigits === 0 && trailingDot && nonInt)
      ? "."
      : "";

  return { int, dot, frac: frac ?? "", suffix };
};
