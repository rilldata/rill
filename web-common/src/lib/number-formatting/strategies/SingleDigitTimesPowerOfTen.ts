import {
  orderOfMagnitude,
  orderOfMagnitudeEng,
  formatNumWithOrderOfMag,
} from "../utils/format-with-order-of-magnitude";
import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";
import {
  FormatterOptionsCommon,
  NumberParts,
  Formatter,
  NumberKind,
  FormatterOptionsIntTimesPowerOfTenStrategy,
} from "../humanizer-types";
import { numberPartsToString } from "../utils/number-parts-utils";

/**
 * detects whether a number is within a floating point error
 * of a single digit multiple of a power of ten.
 */
export const closeToIntTimesPowerOfTen = (x: number) =>
  Math.abs(
    x / 10 ** orderOfMagnitude(x) - Math.round(x / 10 ** orderOfMagnitude(x)),
  ) < 1e-6;

/**
 * This formatter handles numbers that MUST BE (*) a single digit
 * integer multiple of a power of ten, such as the output of
 * d3 scale ticks.
 *
 * Valid examples: 0.000040000, 1, 800, 6000, 5000000
 *
 * Invalid examples: 0.00004300, -12000, 180000, 503000
 *
 * (*) CAVEAT: Because of floating point errors, we accept numbers
 * that are very close to a single digit multiple of a power of ten.
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
export class SingleDigitTimesPowerOfTenFormatter implements Formatter {
  options: FormatterOptionsCommon & FormatterOptionsIntTimesPowerOfTenStrategy;
  initialSample: number[];

  constructor(
    sample: number[],
    options: FormatterOptionsCommon &
      FormatterOptionsIntTimesPowerOfTenStrategy,
  ) {
    this.options = options;
    this.initialSample = sample;
  }

  stringFormat(x: number): string {
    return numberPartsToString(this.partsFormat(x));
  }

  partsFormat(x: number): NumberParts {
    if (typeof x !== "number") {
      // FIXME add these warnings back in when the upstream code is robust enough
      // console.warn(
      //   `Input to SingleDigitTimesPowerOfTenFormatter must be a number, got: ${x}. Returning empty NumberParts.`
      // );
      return { int: "", dot: "", frac: "", suffix: "" };
    }
    const { onInvalidInput } = this.options;

    const isPercent = this.options.numberKind === NumberKind.PERCENT;

    if (isPercent) x = 100 * x;

    let numParts: NumberParts;

    if (x === 0) {
      numParts = { int: "0", dot: "", frac: "", suffix: "" };
    } else {
      if (onInvalidInput === "consoleWarn" || onInvalidInput === "throw") {
        // valid inputs must already be close to a single digit
        //  integer multiple of a power of ten
        if (!closeToIntTimesPowerOfTen(x)) {
          const msg = `received invalid input for SingleDigitTimesPowerOfTenFormatter: ${x}`;
          if (onInvalidInput === "consoleWarn") {
            console.warn(msg);
          } else {
            throw new Error(msg);
          }
        }
      }
      // for this strategy, we NEVER use the trailing dot or pad with zeros
      const useTrailingDot = false;
      const padWithZeros = false;

      const magE = orderOfMagnitudeEng(x);
      numParts = formatNumWithOrderOfMag(
        x,
        magE,
        0,
        padWithZeros,
        useTrailingDot,
      );

      numParts.suffix = shortScaleSuffixIfAvailableForStr(numParts.suffix);

      if (this.options.upperCaseEForExponent !== true) {
        numParts.suffix = numParts.suffix.replace("E", "e");
      }
    }

    if (this.options.numberKind === NumberKind.DOLLAR) {
      numParts.currencySymbol = "$";
    } else if (this.options.numberKind === NumberKind.EURO) {
      numParts.currencySymbol = "€";
    }
    if (this.options.numberKind === NumberKind.PERCENT) {
      numParts.percent = "%";
    }

    return numParts;
  }
}
