import {
  Formatter,
  FormatterOptionsCommon,
  FormatterRangeSpecsStrategy,
  NumberKind,
  NumberParts,
  RangeFormatSpec,
} from "../humanizer-types";
import { countDigits, countNonZeroDigits } from "../utils/count-digits";
import {
  formatNumWithOrderOfMag,
  orderOfMagnitudeEng,
} from "../utils/format-with-order-of-magnitude";
import { numberPartsToString } from "../utils/number-parts-utils";
import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";

const formatWithRangeSpec = (x: number, spec: RangeFormatSpec): NumberParts => {
  const baseMag = spec.baseMagnitude ?? orderOfMagnitudeEng(spec.minMag);
  const padWithInsignificantZeros =
    spec.padWithInsignificantZeros === undefined
      ? true
      : spec.padWithInsignificantZeros;
  const useTrailingDot =
    spec.useTrailingDot === undefined ? true : spec.useTrailingDot;
  return formatNumWithOrderOfMag(
    x,
    baseMag,
    spec.maxDigitsRight,
    padWithInsignificantZeros,
    useTrailingDot
  );
};

const numberPartsValidForRangeSpec = (
  parts: NumberParts,
  spec: RangeFormatSpec
): boolean => {
  const maxDigitsLeft = spec.maxDigitsLeft ?? 3;
  return (
    countDigits(parts.int) <= maxDigitsLeft &&
    countDigits(parts.frac) <= spec.maxDigitsRight
  );
};

const numPartsNotZero = (parts: NumberParts): boolean => {
  return (
    countNonZeroDigits(parts.int) > 0 || countNonZeroDigits(parts.frac) > 0
  );
};

export class PerRangeFormatter implements Formatter {
  options: FormatterOptionsCommon & FormatterRangeSpecsStrategy;
  initialSample: number[];

  // FIXME: we can add this back in if we want to implement
  // alignment. If we decide that we don't want to pursue that,
  // we can remove this commented code
  // largestPossibleNumberStringParts: NumberParts;
  // maxPxWidthsSampledSoFar: FormatterWidths;
  // maxCharWidthsSampledSoFar: FormatterWidths;
  // largestPossibleNumberStringParts: NumberParts;

  constructor(
    sample: number[],
    options: FormatterRangeSpecsStrategy & FormatterOptionsCommon
  ) {
    this.options = options;

    // sort ranges from small to large by lower bound
    this.options.rangeSpecs = this.options.rangeSpecs.sort(
      (a, b) => a.minMag - b.minMag
    );

    // Throw an error if any of the ranges do not have min<sup
    this.options.rangeSpecs.forEach((r) => {
      if (r.minMag >= r.supMag) {
        throw new Error(
          `invalid range: min ${r.minMag} is not strictly less than sup ${r.supMag}`
        );
      }
    });

    // Throw an error if the ranges overlap
    for (let i = 0; i < this.options.rangeSpecs.length - 1; i++) {
      const range1 = this.options.rangeSpecs[i];
      const range2 = this.options.rangeSpecs[i + 1];
      if (range1.supMag > range2.minMag) {
        throw new Error(
          `Ranges must not overlap. range 1 sup = ${range1.supMag} is greater than range 2 min = ${range2.minMag}`
        );
      }
    }

    // Throw an error if there is a gap in coverage overlap
    for (let i = 0; i < this.options.rangeSpecs.length - 1; i++) {
      const range1 = this.options.rangeSpecs[i];
      const range2 = this.options.rangeSpecs[i + 1];
      if (range1.supMag !== range2.minMag) {
        throw new Error(
          `Gaps are not allowed between formatter ranges: range 1 sup = ${range1.supMag} is not equal to range 2 min = ${range2.minMag}`
        );
      }
    }

    this.initialSample = sample;
  }

  stringFormat(x: number): string {
    return numberPartsToString(this.partsFormat(x));
  }

  partsFormat(x: number): NumberParts {
    const { rangeSpecs, defaultMaxDigitsRight } = this.options;

    const isPercent = this.options.numberKind === NumberKind.PERCENT;

    if (isPercent) x = 100 * x;

    let numParts: NumberParts;

    if (x === 0) {
      numParts = { int: "0", dot: "", frac: "", suffix: "" };
    }

    if (numParts === undefined) {
      // Strategy: try to format the number with each spec
      // from smallest OoM to largest, and see whether that
      // range's rules are satisfied and result in a non-zero
      // formatted number.
      for (let i = 0; i < rangeSpecs.length; i++) {
        let spec = rangeSpecs[i];
        numParts = formatWithRangeSpec(x, spec);
        if (
          numberPartsValidForRangeSpec(numParts, spec) &&
          numPartsNotZero(numParts)
        ) {
          // if we have successfully formatted with this spec, we're done...
          break;
        } else {
          // Set numparts back to undefined so and continue
          // through the loop to the next step. If this was the
          // final spec, we'll fall back to the default after
          //  exiting the loop
          numParts = undefined;
        }
      }
    }

    // if numParts is still undefined after trying all specs,
    // use defaults
    if (numParts === undefined) {
      const magE = orderOfMagnitudeEng(x);
      numParts = formatNumWithOrderOfMag(x, magE, defaultMaxDigitsRight, true);
      // Note that if this attempt at formatting results in more than 3
      // digits left of the decimal point, then we must format this
      // number according to the next magnitude up.
      if (countDigits(numParts.int) > 3) {
        numParts = formatNumWithOrderOfMag(
          x,
          magE + 3,
          defaultMaxDigitsRight,
          true
        );
      }
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
}
