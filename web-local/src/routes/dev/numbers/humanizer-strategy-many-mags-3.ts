import { formatNumWithOrderOfMag2 } from "./format-with-order-of-magnitude";
import {
  orderOfMagnitude,
  orderOfMagnitudeEng,
  shortScaleSuffixIfAvailableForStr,
} from "./humanizer-2";
import type { NumberStringParts } from "./number-to-string-formatters";

export const humanizeDefaultStrategy = (
  sample: number[],
  // ordersOfMag: number[],
  // maxOrder: number,
  options
): NumberStringParts[] => {
  // console.log({ options });
  const {
    maxDigitsRight,
    nonIntegerHandling,
    digitTargetPadWithInsignificantZeros,
  } = options;

  let splitStrs: NumberStringParts[] = sample.map((x) => {
    if (x === 0) {
      return { int: "0", dot: "", frac: "", suffix: "" };
    }

    // can the number be shown without suffix within the rules allowed?
    const mag = orderOfMagnitude(x);

    if (mag >= -3 && mag <= 2) {
      // 0.001 to 999.999; format with 3 rhs digits
      return formatNumWithOrderOfMag2(
        x,
        0,
        maxDigitsRight,
        digitTargetPadWithInsignificantZeros,
        nonIntegerHandling === "trailingDot"
      );
    } else if (mag >= 3 && mag <= 5) {
      // 1000 to 999999; format with 0 rhs digits
      return formatNumWithOrderOfMag2(
        x,
        0,
        0,
        false,
        nonIntegerHandling === "trailingDot"
      );
    } else {
      // anything else -- use suffix with 2 RHS digits
      const magE = orderOfMagnitudeEng(x);
      return formatNumWithOrderOfMag2(x, magE, 2, true);
    }
  });

  splitStrs = splitStrs.map((ss) => {
    const suffix = shortScaleSuffixIfAvailableForStr(ss.suffix);
    return { ...ss, ...{ suffix } };
  });

  return splitStrs;
};
