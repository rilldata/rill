import { smallestPrecisionMagnitude } from "./smallest-precision-magnitude";
import { formatNumWithOrderOfMag2 } from "./format-with-order-of-magnitude";
import {
  splitNumStr,
  getSpacingMetadataForSplitStrings,
  getMaxPxWidthsForSplitsStrings,
} from "./num-string-to-aligned-spec";
import type {
  FormatterFactory,
  FormatterSpacingMeta,
  NumberStringParts,
  NumPartPxWidthLookupFn,
} from "./number-to-string-formatters";
import { splitStrsForMagStratMultipleMagsNoAlign } from "./humanizer-strategy-many-mags-2";
import { humanizeDefaultStrategy } from "./humanizer-strategy-many-mags-3";
import { humanizedFormatterFactory } from "@rilldata/web-common/lib/number-formatting/humanizer";
import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import type { FormatterOptionsV1 } from "./formatter-options";

const ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX = {
  0: "",
  3: "k",
  6: "M",
  9: "B",
  12: "T",
  15: "Q",
};

export const shortScaleSuffixIfAvailable = (x: number): string => {
  const suffix = ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX[x];
  if (suffix !== undefined) return suffix;
  return "E" + x;
};

const ORDER_OF_MAG_TEXT_TO_SHORT_SCALE_SUFFIX = {
  E0: "",
  E3: "k",
  E6: "M",
  E9: "B",
  E12: "T",
  E15: "Q",
};
export const shortScaleSuffixIfAvailableForStr = (suffixIn: string): string => {
  const suffix =
    ORDER_OF_MAG_TEXT_TO_SHORT_SCALE_SUFFIX[suffixIn.toUpperCase()];
  if (suffix !== undefined) return suffix;
  return suffixIn;
};

export const orderOfMagnitude = (x) => {
  return Math.floor(Math.log10(Math.abs(x)));
};

export const orderOfMagnitudeEng = (x) => {
  return Math.round(Math.floor(orderOfMagnitude(x) / 3) * 3);
};

export const getOrdersOfMagnitude = (
  sample: number[],
  kind: "engineering" | "scientific" = "scientific"
) => {
  const engFmt = new Intl.NumberFormat("en-US", {
    notation: kind,
  });
  const rawStrings = sample.map(engFmt.format);
  const splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);
  return splitStrs.map((ss) => +ss.suffix.slice(1));
};

export const formatNumWithOrderOfMag = (
  x: number,
  newOrder: number,
  options = { minimumFractionDigits: 3, maximumFractionDigits: 3 }
): NumberStringParts => {
  const [int, frac] = Intl.NumberFormat("en-US", options)
    .format(x / 10 ** newOrder)
    .split(".");
  const dot: "." | "" = ".";

  // if (int === undefined || frac === undefined) {
  //   console.error({ x, int, frac, newOrder });
  // }

  const splitStr = { int, dot, frac: frac ?? "", suffix: "E" + newOrder };

  return splitStr;
};

// window.formatNumWithOrderOfMag = formatNumWithOrderOfMag;

export const thousandthsNumAsDecimalNumParts = (
  x: number,
  maximumFractionDigits = 6,
  padZero = false
): NumberStringParts => {
  // const minimumFractionDigits = padZero ? maximumFractionDigits : 3;

  const orderOfMag = new Intl.NumberFormat("en-US", {
    notation: "engineering",
  })
    .format(x)
    .slice(-3);

  if (orderOfMag !== "E-3") {
    throw new Error(
      `thousandthsNumAsDecimalNumParts only valid for numbers with engineering order of magnitude E-3, got ${x}`
    );
  }

  const formatter = new Intl.NumberFormat("en-US", {
    minimumFractionDigits: padZero ? maximumFractionDigits : 1,
    maximumFractionDigits,
  });

  const [int, frac] = formatter.format(x).split(".");

  return { int, dot: ".", frac, suffix: "" };
};

const splitStrsForMagStratLargestWithDigitsTarget = (
  sample: number[],
  options
): NumberStringParts[] => {
  const { digitTarget, specialDecimalHandling } = options;
  // console.log({ digitTarget });
  const magnitudes = getOrdersOfMagnitude(sample, "scientific");
  const maxMag = Math.max(...magnitudes);
  // if any number is not an integer, may need to reserve one digit
  // after the decimal point to indicate non-integers
  const allAreIntegers = sample.reduce(
    (allSoFar, x) => allSoFar && Number.isInteger(x),
    true
  );

  // Plain integers of reasonable size
  if (0 <= maxMag && maxMag < digitTarget && allAreIntegers) {
    // can just show the plain integers
    const formatter = new Intl.NumberFormat("en-US");
    return sample
      .map((x) => formatter.format(x).replace(",", ""))
      .map(splitNumStr);
  }

  let digitsNeededAfterDecimalForE0 = 1;

  if (
    specialDecimalHandling === "alwaysTwoDigits" ||
    specialDecimalHandling === "neverOneDigit"
  ) {
    digitsNeededAfterDecimalForE0 = 2;
    // NOTE: if the E0 RHS digit adjustment is triggered for this number,
    // then we also need to make sure special padding is applied
  }

  // numbers that can be shown as E0 within digit budget
  if (
    // non-integers of reasonable magnitudes
    (0 <= maxMag &&
      maxMag < digitTarget - digitsNeededAfterDecimalForE0 &&
      !allAreIntegers) ||
    // fractions of reasonable magnitudes
    (0 >= maxMag && maxMag >= -digitTarget)
  ) {
    // if the numbers are not all integers, but the maximum
    // magnitude is such that they'd fit withing the digit
    // target allowing 1 digit after the decimal point,
    // can still use simple formatting, without suffix
    // just need the right number of digits

    const splitStrs = sample.map((x) => {
      // per-number adjustment for specialDecimalHandling
      const minMag = smallestPrecisionMagnitude(x);
      let fracDigits = digitTarget - maxMag - 1;
      let padForE0DecimalHandling = false;
      if (
        specialDecimalHandling === "alwaysTwoDigits" ||
        (minMag === -1 && specialDecimalHandling === "neverOneDigit")
      ) {
        console.log({ x, minMag, specialDecimalHandling });
        fracDigits = 2;
        padForE0DecimalHandling = true;
      }

      console.log({ x, minMag, specialDecimalHandling, fracDigits });

      return formatNumWithOrderOfMag2(
        x,
        0,
        fracDigits,
        options.digitTargetPadWithInsignificantZeros || padForE0DecimalHandling
      );
    });
    splitStrs.forEach((ss) => {
      ss.suffix = "";
    });
    return splitStrs;
  }

  // // FIXME add "minNonzeroDigits" option for this case
  // // fractional number with reasonable number of digits
  // if (0 >= maxMag && maxMag >= -digitTarget) {
  //   // if maxMag represents a fraction that can be shown within
  //   // digitTarget digits of the decimal point,
  //   // use simple formatting without suffix
  //   // let formatter = new Intl.NumberFormat("en-US", {
  //   //   minimumFractionDigits: digitTarget,
  //   //   maximumFractionDigits: digitTarget,
  //   // });
  //   // return sample
  //   //   .map((x) => formatter.format(x).replace(",", ""))
  //   //   .map(splitNumStr);

  //   let splitStrs = sample.map((x) => {
  //     const minMag = smallestPrecisionMagnitude(x);
  //     let fracDigits = digitTarget - maxMag - 1;
  //     let padForE0DecimalHandling = false;
  //     if (
  //       specialDecimalHandling === "alwaysTwoDigits" ||
  //       (minMag === -1 && specialDecimalHandling === "neverOneDigit")
  //     ) {
  //       console.log({ x, minMag, specialDecimalHandling });
  //       fracDigits = 2;
  //       padForE0DecimalHandling = true;
  //     }

  //     console.log({ x, minMag, specialDecimalHandling, fracDigits });

  //     return formatNumWithOrderOfMag2(
  //       x,
  //       0,
  //       fracDigits,
  //       options.digitTargetPadWithInsignificantZeros || padForE0DecimalHandling
  //     );
  //   });
  //   splitStrs.forEach((ss) => {
  //     ss.suffix = "";
  //   });
  //   return splitStrs;
  // }

  // At this point, the largest magnitude represents
  // either a tiny infinitesimal, or a large number.
  // Use standard 3 order of mag groupings and a suffix.
  const maxMagEng = Math.floor(maxMag / 3) * 3;
  const intDigits = maxMag - maxMagEng + 1;
  const fracDigits = digitTarget - intDigits;
  // console.log({ intDigits, fracDigits });
  const splitStrs = sample.map((x) =>
    formatNumWithOrderOfMag(x, maxMagEng, {
      minimumFractionDigits: fracDigits,
      maximumFractionDigits: fracDigits,
    })
  );
  const maxOrderSuffix = shortScaleSuffixIfAvailable(maxMagEng);
  splitStrs.forEach((ss) => {
    ss.suffix = maxOrderSuffix;
  });
  return splitStrs;
};

const splitStrsForMagStratLargest = (
  sample: number[],
  ordersOfMag: number[],
  maxOrder: number,
  options
): NumberStringParts[] => {
  let maxOrderSuffix: string;

  let splitStrs: NumberStringParts[];

  if (options.usePlainNumsForThousands && maxOrder === 3) {
    // if top magnitude is e3 (thousands) AND ALL ARE INTEGERS, can just show 6 digits of integer parts
    const decimals = options.usePlainNumsForThousandsOneDecimal ? 1 : 0;
    const formatter = new Intl.NumberFormat("en-US", {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    });

    splitStrs = sample
      .map((x) => formatter.format(x).replace(",", ""))
      .map(splitNumStr)
      .map((splitStr, i) => {
        if (Number.isInteger(sample[i])) {
          splitStr.frac = "";
          splitStr.dot = "";
        }
        return splitStr;
      });
    maxOrderSuffix = "";
  } else if (options.usePlainNumForThousandths && maxOrder === -3) {
    // const formatter = new Intl.NumberFormat("en-US", {
    //   minimumFractionDigits: options.usePlainNumForThousandthsPadZeros ? 6 : 1,
    //   maximumFractionDigits: 6,
    // });

    // splitStrs = sample.map((x) => formatter.format(x)).map(splitNumStr);

    const minimumFractionDigits = options.usePlainNumForThousandthsPadZeros
      ? 6
      : 1;
    // maximumFractionDigits: 6,

    splitStrs = sample.map((x) =>
      formatNumWithOrderOfMag(x, 0, {
        minimumFractionDigits,
        maximumFractionDigits: 6,
      })
    );

    maxOrderSuffix = "";
  } else {
    splitStrs = sample.map((x) => formatNumWithOrderOfMag(x, maxOrder));
    maxOrderSuffix = shortScaleSuffixIfAvailable(maxOrder);
  }

  splitStrs.forEach((ss) => {
    ss.suffix = maxOrderSuffix;
  });

  return splitStrs;
};

const splitStrsForMagStratUnlimited = (
  sample: number[],
  ordersOfMag: number[],
  maxOrder: number,
  options
): NumberStringParts[] => {
  const engFmt = new Intl.NumberFormat("en-US", {
    notation: "engineering",
    minimumFractionDigits: 3,
  });
  const rawStrings = sample.map(engFmt.format);
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  splitStrs = splitStrs.map((ss, i) => {
    if (options.truncateThousandths && ordersOfMag[i] === -3) {
      return thousandthsNumAsDecimalNumParts(sample[i], 3);
    }

    if (
      options.truncateTinyOrdersIfBigOrderExists &&
      ordersOfMag[i] < -3 &&
      maxOrder >= -3
    ) {
      const ss = formatNumWithOrderOfMag(sample[i], 0);
      ss.suffix = "";
      return ss;
    }

    const suffix = shortScaleSuffixIfAvailable(ordersOfMag[i]);
    return { ...ss, ...{ suffix } };
  });

  return splitStrs;
};

export const humanized2FormatterFactory: FormatterFactory = (
  sample: number[],
  pxWidthLookup: NumPartPxWidthLookupFn,
  options: FormatterOptionsV1
) => {
  const range = { max: Math.max(...sample), min: Math.min(...sample) };

  const engFmt = new Intl.NumberFormat("en-US", {
    notation: "engineering",
    minimumFractionDigits: 3,
  });
  const rawStrings = sample.map(engFmt.format);
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  const ordersOfMag = splitStrs.map((ss) => +ss.suffix.slice(1));

  // omit exact zeros when calculating orders of magnitude
  const orderOfMagNoZeros = ordersOfMag.filter((_, i) => sample[i] != 0);
  const maxOrder = Math.max(...orderOfMagNoZeros);

  splitStrs.forEach((ss, i) => {
    const suff = ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX[ordersOfMag[i]];
    if (suff !== undefined) ss.suffix = suff;
  });

  splitStrs.forEach((ss) => {
    if (ss.suffix === undefined) console.error("bad suffix pre", ss);
  });

  switch (options.magnitudeStrategy) {
    case "largest":
      splitStrs = splitStrsForMagStratLargest(
        sample,
        ordersOfMag,
        maxOrder,
        options
      );
      break;
    case "unlimited":
      splitStrs = splitStrsForMagStratUnlimited(
        sample,
        ordersOfMag,
        maxOrder,
        options
      );
      break;

    case "largestWithDigitTarget":
      splitStrs = splitStrsForMagStratLargestWithDigitsTarget(sample, options);
      // console.log("splitStrs", splitStrs);
      break;

    case "unlimitedDigitTarget":
      splitStrs = splitStrsForMagStratMultipleMagsNoAlign(sample, options);
      // console.log("splitStrs", splitStrs);
      break;

    case "defaultStrategyProposal-2023-03-02":
      splitStrs = humanizeDefaultStrategy(sample, options);
      // console.log("splitStrs", splitStrs);
      break;

    case "defaultStrategy":
      {
        let numberKind = NumberKind.ANY;
        switch (options.formattingUnits) {
          case "$":
            numberKind = NumberKind.DOLLAR;
            break;
          case "%":
            numberKind = NumberKind.PERCENT;
            break;
          default:
            break;
        }

        const formatter = humanizedFormatterFactory(sample, {
          strategy: "default",
          numberKind,
        });
        splitStrs = sample.map((x) => formatter.partsFormat(x));
      }
      // console.log("splitStrs", splitStrs);
      break;

    default:
      break;
  }

  splitStrs.forEach((ss) => {
    if (ss.suffix === undefined) console.error("bad suffix post", ss);
    // FIXME: add concept of "prefix" instead of this hack
    if (options.formattingUnits === "$") {
      ss.int = ss.int[0] === "-" ? "-$" + ss.int.slice(1) : "$" + ss.int;
    }
    if (options.formattingUnits === "%") {
      ss.suffix += "%";
    }
  });

  const spacing: FormatterSpacingMeta =
    getSpacingMetadataForSplitStrings(splitStrs);
  const maxPxWidth = getMaxPxWidthsForSplitsStrings(splitStrs, pxWidthLookup);

  return (x: number) => {
    const i = sample.findIndex((h) => h === x);
    return {
      number: x,
      rawStr: rawStrings[i],
      splitStr: splitStrs[i],
      spacing,
      range,
      maxPxWidth,
    };
  };
};
