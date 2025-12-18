import {
  type Formatter,
  type FormatterOptionsCommon,
  type FormatterRangeSpecsStrategy,
  type LocaleConfig,
  NumberKind,
  type NumberParts,
  type RangeFormatSpec,
} from "../humanizer-types";
import { countDigits, countNonZeroDigits } from "../utils/count-digits";
import {
  formatNumWithOrderOfMag,
  orderOfMagnitude,
  orderOfMagnitudeEng,
} from "../utils/format-with-order-of-magnitude";
import { numberPartsToString } from "../utils/number-parts-utils";
import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";

const formatWithRangeSpec = (
  x: number,
  spec: RangeFormatSpec,
  locale?: LocaleConfig,
): NumberParts => {
  if (spec.overrideValue !== undefined) {
    return { ...spec.overrideValue, locale };
  }
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
    useTrailingDot,
    false,
    locale,
  );
};

const numberPartsValidForRangeSpec = (
  parts: NumberParts,
  spec: RangeFormatSpec,
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
  locale?: LocaleConfig;

  constructor(
    options: FormatterRangeSpecsStrategy & FormatterOptionsCommon,
    locale?: LocaleConfig,
  ) {
    this.options = options;
    this.locale = locale;

    // sort ranges from small to large by lower bound
    this.options.rangeSpecs = this.options.rangeSpecs.sort(
      (a, b) => a.minMag - b.minMag,
    );

    // Throw an error if any of the ranges do not have min<sup
    this.options.rangeSpecs.forEach((r) => {
      if (r.minMag >= r.supMag) {
        throw new Error(
          `invalid range: min ${r.minMag} is not strictly less than sup ${r.supMag}`,
        );
      }
    });

    // Throw an error if the ranges overlap
    for (let i = 0; i < this.options.rangeSpecs.length - 1; i++) {
      const range1 = this.options.rangeSpecs[i];
      const range2 = this.options.rangeSpecs[i + 1];
      if (range1.supMag > range2.minMag) {
        throw new Error(
          `Ranges must not overlap. range 1 sup = ${range1.supMag} is greater than range 2 min = ${range2.minMag}`,
        );
      }
    }

    // Throw an error if there is a gap in coverage overlap
    for (let i = 0; i < this.options.rangeSpecs.length - 1; i++) {
      const range1 = this.options.rangeSpecs[i];
      const range2 = this.options.rangeSpecs[i + 1];
      if (range1.supMag !== range2.minMag) {
        throw new Error(
          `Gaps are not allowed between formatter ranges: range 1 sup = ${range1.supMag} is not equal to range 2 min = ${range2.minMag}`,
        );
      }
    }
  }

  stringFormat(x: number): string {
    return numberPartsToString(this.partsFormat(x));
  }

  partsFormat(x: number): NumberParts {
    const { rangeSpecs, defaultMaxDigitsRight } = this.options;

    const isPercent = this.options.numberKind === NumberKind.PERCENT;

    if (isPercent) x = 100 * x;

    let numParts: NumberParts | undefined = undefined;

    if (x === 0) {
      numParts = {
        int: "0",
        dot: "",
        frac: "",
        suffix: "",
        locale: this.locale,
      };
    }

    if (numParts === undefined) {
      // Strategy: try to format the number with each spec
      // from smallest OoM to largest, and see whether that
      // range's rules are satisfied and result in a non-zero
      // formatted number.
      for (let i = 0; i < rangeSpecs.length; i++) {
        const spec = rangeSpecs[i];
        if (spec.overrideValue !== undefined) {
          const OoM = orderOfMagnitude(x);
          if (OoM >= spec.minMag && OoM < spec.supMag) {
            numParts = { ...spec.overrideValue, locale: this.locale };
            break;
          }
        }
        numParts = formatWithRangeSpec(x, spec, this.locale);
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
      const hasSmallFraction = magE <= -3;

      const maxDigitsLeft = hasSmallFraction ? 1 : 3;
      numParts = formatNumWithOrderOfMag(
        x,
        magE,
        defaultMaxDigitsRight,
        true,
        false,
        false,
        this.locale,
      );
      // Note that if this attempt at formatting results in more
      // digits left of the decimal point than maxDigitsLeft, then we must format this
      // number according to the next magnitude up. We'll attempt this up to 3 times
      // to find a suitable formatting, incrementing the magnitude each time.
      let attempts = 0;
      while (countDigits(numParts.int) > maxDigitsLeft && attempts < 3) {
        numParts = formatNumWithOrderOfMag(
          x,
          magE + maxDigitsLeft + attempts,
          defaultMaxDigitsRight,
          true,
          false,
          false,
          this.locale,
        );
        attempts++;
      }
    }

    numParts.suffix = shortScaleSuffixIfAvailableForStr(numParts.suffix);

    if (this.options.upperCaseEForExponent !== true) {
      numParts.suffix = numParts.suffix.replace("E", "e");
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
