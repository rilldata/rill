import {
  orderOfMagnitude,
  orderOfMagnitudeEng,
  formatNumWithOrderOfMag,
} from "../utils/format-with-order-of-magnitude";
import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";
import {
  FormatterOptionsCommon,
  FormatterOptionsDefaultStrategy,
  FormatterWidths,
  NumberParts,
  Formatter,
  NumberKind,
  FormatterOptionsIntTimesPowerOfTenStrategy,
} from "../humanizer-types";
import { numberPartsToString } from "../utils/number-parts-utils";
import { smallestPrecisionMagnitude } from "../utils/smallest-precision-magnitude";

export class IntTimesPowerOfTenFormatter implements Formatter {
  options: FormatterOptionsCommon & FormatterOptionsIntTimesPowerOfTenStrategy;
  initialSample: number[];

  maxPxWidthsSampledSoFar: FormatterWidths;
  maxCharWidthsSampledSoFar: FormatterWidths;

  largestPossibleNumberStringParts: NumberParts;

  constructor(
    sample: number[],
    options: FormatterOptionsCommon & FormatterOptionsIntTimesPowerOfTenStrategy
  ) {
    this.options = options;
    this.initialSample = sample;

    this.largestPossibleNumberStringParts = {
      neg: "-",
      dollar: options.numberKind === NumberKind.DOLLAR ? "$" : undefined,
      int: "999",
      dot: "",
      frac: "",
      suffix: "e-324",
      percent: options.numberKind === NumberKind.PERCENT ? "%" : undefined,
    };
  }

  stringFormat(x: number): string {
    return numberPartsToString(this.partsFormat(x));
  }

  partsFormat(x: number): NumberParts {
    const { onInvalidInput } = this.options;

    const isCurrency = this.options.numberKind === NumberKind.DOLLAR;
    const isPercent = this.options.numberKind === NumberKind.PERCENT;

    if (isPercent) x = 100 * x;

    let numParts: NumberParts;

    if (x === 0) {
      numParts = { int: "0", dot: "", frac: "", suffix: "" };
    } else {
      if (onInvalidInput === "consoleWarn" || onInvalidInput === "throw") {
        // valid inputs must have only one digit of precision
        // -- i.e, the leading OoM must match the OoM of the most precise digit
        if (orderOfMagnitude(x) !== smallestPrecisionMagnitude(x)) {
          const msg = `recieved invalid input for IntTimesPowerOfTenFormatter: ${x}`;
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
        useTrailingDot
      );

      numParts.suffix = shortScaleSuffixIfAvailableForStr(numParts.suffix);

      if (this.options.upperCaseEForExponent !== true) {
        numParts.suffix = numParts.suffix.replace("E", "e");
      }
    }

    if (isCurrency) {
      numParts.dollar = "$";
    }
    if (this.options.numberKind === NumberKind.PERCENT) {
      numParts.percent = "%";
    }

    return numParts;
  }

  // FIXME? -- will be needed if we want alignment
  // updateMaxWidthsSample(x: number) {}

  // maxPxWidthsSampled(): FormatterWidths;
  // maxPxWidthsPossible(): FormatterWidths;

  // maxCharWidthsSampled(): FormatterWidths;
  // maxCharWidthsPossible(): FormatterWidths;
}
