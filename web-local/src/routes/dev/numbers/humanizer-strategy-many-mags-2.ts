import { formatNumWithOrderOfMag2 } from "./format-with-order-of-magnitude";
import {
  shortScaleSuffixIfAvailable,
  thousandthsNumAsDecimalNumParts,
  orderOfMagnitude,
  orderOfMagnitudeEng,
  shortScaleSuffixIfAvailableForStr,
} from "./humanizer-2";
import { splitNumStr } from "./num-string-to-aligned-spec";
import type { NumberStringParts } from "./number-to-string-formatters";
import { smallestPrecisionMagnitude } from "./smallest-precision-magnitude";

export const splitStrsForMagStratMultipleMagsNoAlign = (
  sample: number[],
  // ordersOfMag: number[],
  // maxOrder: number,
  options
): NumberStringParts[] => {
  // console.log({ options });
  const {
    maxTotalDigits,
    maxDigitsLeft,
    maxDigitsRight,
    minDigitsNonzero,
    useMaxDigitsRightIfSuffix,
    maxDigitsRightIfSuffix,
    nonIntegerHandling,
    digitTargetPadWithInsignificantZeros,
    specialDecimalHandling,
  } = options;

  let splitStrs: NumberStringParts[] = sample.map((x) => {
    if (x === 0) {
      return { int: "0", dot: "", frac: "", suffix: "" };
    }

    // can the number be shown without suffix within the rules allowed?
    const mag = orderOfMagnitude(x);
    const minMag = smallestPrecisionMagnitude(x);
    const digits = mag - (minMag < 0 ? minMag : 0) + 1;

    let RHSDigits;
    let LHSDigits;

    let padForE0DecimalHandling = false;

    if (mag < 0) {
      // first consider fractional numbers.

      // In this case, all  digits will be RHS, so we just need
      // to check whether formatting with E0 allows enough non-zero digits.

      // get the number of RHS digits
      RHSDigits = Math.min(maxTotalDigits - 1, maxDigitsRight); // -1 for leading 0

      // apply special decimal handling for e0
      if (
        specialDecimalHandling === "alwaysTwoDigits" ||
        (minMag === -1 && specialDecimalHandling === "neverOneDigit")
      ) {
        RHSDigits = 2;
        // NOTE: if the E0 RHS digit adjustment is triggered for this number,
        // then we also need to make sure special padding is applied
        padForE0DecimalHandling = true;
      }

      // do the formatting
      let ss = formatNumWithOrderOfMag2(
        x,
        0,
        RHSDigits,
        digitTargetPadWithInsignificantZeros || padForE0DecimalHandling
      );

      // see how many signif digits this number has:
      const numSignifDigits = -minMag + mag;
      // the representation needs to have at least minDigitsNonzero,
      // unless the number doesn't have that many signif digits to start/
      const thisMinDigitsNonzero = Math.min(minDigitsNonzero, numSignifDigits);

      // if this representation has enough significant digits, return it...
      let nonZeroDigits = ss.frac
        .split("")
        .filter((char) => char !== "0").length;
      if (nonZeroDigits >= thisMinDigitsNonzero) return ss;

      // ...otherwise, format with order of mag
      const magE = orderOfMagnitudeEng(x);
      LHSDigits = mag - magE + 1;
      if (useMaxDigitsRightIfSuffix) {
        RHSDigits = maxDigitsRightIfSuffix;
      } else {
        RHSDigits = Math.min(maxTotalDigits - LHSDigits, maxDigitsRight);
      }
      return formatNumWithOrderOfMag2(
        x,
        magE,
        RHSDigits,
        digitTargetPadWithInsignificantZeros
      );
    } else {
      // now considering numbers with an integer part (ie mag >= 0)

      // check whether we need to reserve a digit for
      // a fractional part
      const isInt = Number.isInteger(x);
      const nonIntReserveDigit =
        !isInt && nonIntegerHandling === "oneDigit" ? 1 : 0;

      // we can format this number without a magnitude if:
      // (a) it's magniture fits within our maxTotalDigits
      // budget, including handling of a nonInt digit
      // (b) it's magniture fits within our maxDigitsLeft
      // budget, including handling of a nonInt digit

      // try formatting the number at e0...
      LHSDigits = Math.min(mag + 1, maxTotalDigits, maxDigitsLeft);
      RHSDigits = Math.min(maxTotalDigits - LHSDigits, maxDigitsRight);

      // apply special decimal handling for e0
      if (
        specialDecimalHandling === "alwaysTwoDigits" ||
        (minMag === -1 && specialDecimalHandling === "neverOneDigit")
      ) {
        RHSDigits = 2;
        // NOTE: if the E0 RHS digit adjustment is triggered for this number,
        // then we also need to make sure special padding is applied
        padForE0DecimalHandling = true;
        console.log({ x, RHSDigits });
      }

      let ss = formatNumWithOrderOfMag2(
        x,
        0,
        RHSDigits,
        digitTargetPadWithInsignificantZeros || padForE0DecimalHandling,
        nonIntegerHandling === "trailingDot"
      );

      // At this point, by construction, RHSDigits<= maxDigitsRight;
      // Need to check that
      // (a) maxDigitsLeft constraint is satisfied
      // (b) nonIntegerHandling rules satisfied
      const minRhsDigits = !isInt && nonIntegerHandling === "oneDigit" ? 1 : 0;

      if (ss.int.length <= maxDigitsLeft && ss.frac.length >= minRhsDigits) {
        return ss;
      }

      // ...otherwise, format with order of mag
      const magE = orderOfMagnitudeEng(x);
      LHSDigits = mag - magE + 1;
      if (useMaxDigitsRightIfSuffix) {
        RHSDigits = maxDigitsRightIfSuffix;
      } else {
        RHSDigits = Math.min(maxTotalDigits - LHSDigits, maxDigitsRight);
      }
      return formatNumWithOrderOfMag2(
        x,
        magE,
        RHSDigits,
        digitTargetPadWithInsignificantZeros
      );

      // if (mag >= 0 && mag < maxTotalDigits + nonIntReserveDigit) {
      //   const fracDigitsNeeded = maxTotalDigits - mag;
      //   const;
      // }

      //   const fracDot = !isInt && nonIntegerHandling !== "noSpecial" ? "." : "";
    }
  });

  splitStrs = splitStrs.map((ss, i) => {
    let suffix = shortScaleSuffixIfAvailableForStr(ss.suffix);
    return { ...ss, ...{ suffix } };
  });

  return splitStrs;
};
