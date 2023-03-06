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
} from "../humanizer-types";
import { numberPartsToString } from "../utils/number-parts-utils";

export const humanizeDefaultStrategy = (
  sample: number[],
  options: FormatterOptionsCommon & FormatterOptionsDefaultStrategy
): NumberParts[] => {
  const {
    // number of RHS digits for x s.t.: 1e-3 <= x < 1e6
    maxDigitsRightSmallNums,
    // number of RHS digits for numbers rendered with a suffix
    maxDigitsRightSuffixNums,
    padWithInsignificantZeros,
  } = options;

  // for default strategy, we'll always use the trailing dot
  const useTrailingDot = true;

  let numPartsArr: NumberParts[] = sample.map((x) => {
    if (x === 0) {
      return { int: "0", dot: "", frac: "", suffix: "" };
    }

    // can the number be shown without suffix within the rules allowed?
    const mag = orderOfMagnitude(x);

    if (mag >= -3 && mag <= 2) {
      // 0.001 to 999.999; format with 3 rhs digits
      return formatNumWithOrderOfMag(
        x,
        0,
        maxDigitsRightSmallNums,
        padWithInsignificantZeros,
        useTrailingDot
      );
    } else if (mag >= 3 && mag <= 5) {
      // 1000 to 999999; format with 0 rhs digits
      return formatNumWithOrderOfMag(x, 0, 0, false, useTrailingDot);
    } else {
      // anything else -- use suffix with maxDigitsRightSuffixNums
      const magE = orderOfMagnitudeEng(x);
      return formatNumWithOrderOfMag(x, magE, maxDigitsRightSuffixNums, true);
    }
  });

  numPartsArr = numPartsArr.map((ss, i) => {
    let suffix = shortScaleSuffixIfAvailableForStr(ss.suffix);
    return { ...ss, ...{ suffix } };
  });

  return numPartsArr;
};

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

export class DefaultHumanizer implements Formatter {
  options: FormatterOptionsCommon & FormatterOptionsDefaultStrategy;
  initialSample: number[];

  maxPxWidthsSampledSoFar: FormatterWidths;
  maxCharWidthsSampledSoFar: FormatterWidths;

  largestPossibleNumberStringParts: NumberParts;

  constructor(
    sample: number[],
    options: FormatterOptionsCommon & FormatterOptionsDefaultStrategy
  ) {
    this.options = options;
    this.initialSample = sample;

    const { maxDigitsRightSmallNums, maxDigitsRightSuffixNums } = this.options;

    this.largestPossibleNumberStringParts = {
      neg: "-",
      dollar: options.numberKind === NumberKind.DOLLAR ? "$" : undefined,
      int: "999999",
      dot: ".",
      frac: "0".repeat(
        Math.max(maxDigitsRightSmallNums, maxDigitsRightSuffixNums)
      ),
      suffix: "e-324",
      percent: options.numberKind === NumberKind.PERCENT ? "%" : undefined,
    };
  }

  stringFormat(x: number): string {
    return numberPartsToString(this.partsFormat(x));
  }

  partsFormat(x: number): NumberParts {
    const {
      // number of RHS digits for x s.t.: 1e-3 <= x < 1e6
      maxDigitsRightSmallNums,
      // number of RHS digits for numbers rendered with a suffix
      maxDigitsRightSuffixNums,
      padWithInsignificantZeros,
    } = this.options;

    // for default strategy, we'll always use the trailing dot
    const useTrailingDot = true;

    let numParts: NumberParts;

    if (x === 0) {
      numParts = { int: "0", dot: "", frac: "", suffix: "" };
    }

    // can the number be shown without suffix within the rules allowed?
    const mag = orderOfMagnitude(x);

    const isCurrency = this.options.numberKind === NumberKind.DOLLAR;

    if (mag >= -3 && mag <= 2 && !isCurrency) {
      // 0.001 to 999.999 and NOT currency; format with 3 rhs digits
      numParts = formatNumWithOrderOfMag(
        x,
        0,
        maxDigitsRightSmallNums,
        padWithInsignificantZeros,
        useTrailingDot
      );
    } else if (mag >= -2 && mag <= 2 && isCurrency) {
      // 0.01 to 999.999 and IS currency; format with 2 rhs digits
      numParts = formatNumWithOrderOfMag(
        x,
        0,
        2,
        padWithInsignificantZeros,
        useTrailingDot
      );
    } else if (mag >= 3 && mag <= 5) {
      // 1000 to 999999; format with 0 rhs digits
      numParts = formatNumWithOrderOfMag(x, 0, 0, false, useTrailingDot);
    } else {
      // anying else -- use suffix with maxDigitsRightSuffixNums
      const magE = orderOfMagnitudeEng(x);
      numParts = formatNumWithOrderOfMag(
        x,
        magE,
        maxDigitsRightSuffixNums,
        true
      );
    }

    numParts.suffix = shortScaleSuffixIfAvailableForStr(numParts.suffix);

    if (this.options.upperCaseEForExponent !== true) {
      numParts.suffix = numParts.suffix.replace("E", "e");
    }

    if (this.options.numberKind === NumberKind.DOLLAR) {
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
