import {
  orderOfMagnitude,
  orderOfMagnitudeEng,
  formatNumWithOrderOfMag,
} from "../utils/format-with-order-of-magnitude";
import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";
import type {
  FormatterOptionsCommon,
  FormatterOptionsDefaultStrategy,
  NumberStringParts,
} from "../humanizer-types";

export const humanizeDefaultStrategy = (
  sample: number[],
  options: FormatterOptionsCommon & FormatterOptionsDefaultStrategy
): NumberStringParts[] => {
  const {
    // number of RHS digits for x s.t.: 1e-3 <= x < 1e6
    maxDigitsRightSmallNums,
    // number of RHS digits for numbers rendered with a suffix
    maxDigitsRightSuffixNums,
    padWithInSignificantZeros,
  } = options;

  // for default strategy, we'll always use the trailing dot
  const useTrailingDot = true;

  let splitStrs: NumberStringParts[] = sample.map((x) => {
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
        padWithInSignificantZeros,
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

  splitStrs = splitStrs.map((ss, i) => {
    let suffix = shortScaleSuffixIfAvailableForStr(ss.suffix);
    return { ...ss, ...{ suffix } };
  });

  return splitStrs;
};
