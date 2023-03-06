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
  RichFormatNumber,
  NumberKind,
  FormatterOptionsNoneStrategy,
} from "../humanizer-types";
import {
  numberPartsToString,
  numStrToParts,
} from "../utils/number-parts-utils";

// FIXME? -- will be needed if we want alignment
// export const humanizeDefaultStrategyMaxCharWidthsPossible = (
//   options: FormatterOptionsCommon & FormatterOptionsDefaultStrategy
// ): FormatterWidths => {
//   const {
//     // number of RHS digits for x s.t.: 1e-3 <= x < 1e6
//     maxDigitsRightSmallNums,
//     // number of RHS digits for numbers rendered with a suffix
//     maxDigitsRightSuffixNums,
//   } = options;

//   return {
//     // max ever is 8 for e.g. "-$999999"
//     left: 6,
//     // max ever is 1 for "."
//     dot: 1,

//     frac: Math.max(maxDigitsRightSmallNums, maxDigitsRightSuffixNums),

//     // max ever is 6 for e.g. "e-324%"
//     suffix: 6,
//   };
// };

// FIXME? -- will be needed if we want alignment
// export const humanizeDefaultStrategyMaxPxWidthsPossible = (
//   options: FormatterOptionsCommon & FormatterOptionsDefaultStrategy
// ): FormatterWidths => {
//   const {
//     // number of RHS digits for x s.t.: 1e-3 <= x < 1e6
//     maxDigitsRightSmallNums,
//     // number of RHS digits for numbers rendered with a suffix
//     maxDigitsRightSuffixNums,
//   } = options;

//   return {
//     // max ever is 8 for e.g. "-$999999"
//     left: 6,
//     // max ever is 1 for "."
//     dot: 1,

//     frac: Math.max(maxDigitsRightSmallNums, maxDigitsRightSuffixNums),

//     // max ever is 6 for e.g. "e-324%"
//     suffix: 6,
//   };
// };

export class NonFormatter implements Formatter {
  options: FormatterOptionsCommon & FormatterOptionsNoneStrategy;
  initialSample: number[];

  maxPxWidthsSampledSoFar: FormatterWidths;
  maxCharWidthsSampledSoFar: FormatterWidths;

  largestPossibleNumberStringParts: NumberParts;

  constructor(
    sample: number[],
    options: FormatterOptionsCommon & FormatterOptionsNoneStrategy
  ) {
    this.options = options;
    this.initialSample = sample;

    // const { maxDigitsRightSmallNums, maxDigitsRightSuffixNums } = this.options;

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
      })
        .format(x)
        .replaceAll(",", "");
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

  // FIXME? -- will be needed if we want alignment
  // updateMaxWidthsSample(x: number) {}

  // maxPxWidthsSampled(): FormatterWidths;
  // maxPxWidthsPossible(): FormatterWidths;

  // maxCharWidthsSampled(): FormatterWidths;
  // maxCharWidthsPossible(): FormatterWidths;
}
