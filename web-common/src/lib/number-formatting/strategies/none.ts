import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";
import {
  FormatterOptionsCommon,
  NumberParts,
  Formatter,
  NumberKind,
  FormatterOptionsNoneStrategy,
} from "../humanizer-types";
import {
  numberPartsToString,
  numStrToParts,
} from "../utils/number-parts-utils";

export class NonFormatter implements Formatter {
  options: FormatterOptionsCommon & FormatterOptionsNoneStrategy;
  initialSample: number[];

  // FIXME: we can add this back in if we want to implement
  // alignment. If we decide that we don't want to pursue that,
  // we can remove this commented code
  // maxPxWidthsSampledSoFar: FormatterWidths;
  // maxCharWidthsSampledSoFar: FormatterWidths;
  // largestPossibleNumberStringParts: NumberParts;

  constructor(
    sample: number[],
    options: FormatterOptionsCommon & FormatterOptionsNoneStrategy
  ) {
    this.options = options;
    this.initialSample = sample;

    // FIXME: we can add this back in if we want to implement
    // alignment. If we decide that we don't want to pursue that,
    // we can remove this commented code
    // largestPossibleNumberStringParts: NumberParts;

    // this.largestPossibleNumberStringParts = {
    //   neg: "-",
    //   dollar: options.numberKind === NumberKind.DOLLAR ? "$" : undefined,
    //   int: "999999",
    //   dot: ".",
    //   frac: "0".repeat(
    //     Math.max(maxDigitsRightSmallNums, maxDigitsRightSuffixNums)
    //   ),
    //   suffix: "e-324",
    //   percent: options.numberKind === NumberKind.PERCENT ? "%" : undefined,
    // };
  }

  stringFormat(x: number): string {
    return numberPartsToString(this.partsFormat(x));
  }

  partsFormat(x: number): NumberParts {
    let numParts;

    const isCurrency = this.options.numberKind === NumberKind.DOLLAR;
    const isPercent = this.options.numberKind === NumberKind.PERCENT;

    if (isPercent) x = 100 * x;

    if (x === 0) {
      numParts = { int: "0", dot: "", frac: "", suffix: "" };
    } else {
      const str = new Intl.NumberFormat("en", {
        maximumFractionDigits: 20,
      }).format(x);
      numParts = numStrToParts(str);
    }

    numParts.suffix = shortScaleSuffixIfAvailableForStr(numParts.suffix);

    if (this.options.upperCaseEForExponent !== true) {
      numParts.suffix = numParts.suffix.replace("E", "e");
    }

    if (isCurrency) {
      numParts.dollar = "$";
    }
    if (isPercent) {
      numParts.percent = "%";
    }

    return numParts;
  }
}
