import type { LocaleConfig, NumberParts } from "../humanizer-types";
import { smallestPrecisionMagnitude } from "./smallest-precision-magnitude";

export const orderOfMagnitude = (x) => {
  if (x === 0) return 0;
  let mag = Math.floor(Math.log10(Math.abs(x)));
  // having found the order of magnitude, if we divide it
  // out of the number, we should get a number between 1 and 10.
  // However, because of floating point errors, if we get a number
  // very just less than 10, we may have a floating point error,
  // in which we want to bump the order of magnitude up by one.
  //
  // Ex: 0.0009999999999999 has magnitude -4, but if multiply away
  // the magnitude, we get:
  // 0.0009999999999999 * 10**4 = 9.999999999999999
  // -- just less than 10, so we want to bump the magnitude up to -3
  // so that we can formar this as e.g. "1.0e-3"
  if (10 - Math.abs(x) * 10 ** -mag < 1e-8) mag += 1;
  return mag;
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
  trailingDot = false,

  // strip commas from output?
  stripCommas = false,

  // Optional locale configuration for formatting
  locale?: LocaleConfig,
): NumberParts => {
  if (typeof x !== "number") {
    // FIXME add these warnings back in when the upstream code is robust enough
    // console.warn(
    //   `input to formatNumWithOrderOfMag must be a number, got: ${x}. Returning empty NumberParts.`
    // );
    return { int: "", dot: "", frac: "", suffix: "" };
  }

  if (x === Infinity) return { int: "∞", dot: "", frac: "", suffix: "" };
  if (x === -Infinity)
    return { neg: "-", int: "∞", dot: "", frac: "", suffix: "" };
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

  const int_frac = Intl.NumberFormat("en-US", {
    maximumFractionDigits: fractionDigits,
    minimumFractionDigits: fractionDigits,
  })
    .format(x / 10 ** newOrder)
    .split(".");

  let int = int_frac[0];
  const frac = int_frac[1];

  // Always strip commas when locale is provided - we'll apply locale-specific separators later
  if (stripCommas || locale) {
    int = int.replace(/,/g, "");
  }

  // handle negatives
  let neg: "-" | undefined = undefined;
  if (int[0] === "-") {
    int = int.slice(1);
    neg = "-";
  }

  const nonInt = !Number.isInteger(x);

  dot =
    frac !== undefined || (fractionDigits === 0 && trailingDot && nonInt)
      ? "."
      : "";

  return { neg, int, dot, frac: frac ?? "", suffix, locale };
};
