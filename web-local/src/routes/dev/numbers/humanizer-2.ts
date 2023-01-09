import {
  humanizeDataType,
  humanizeGroupValues,
  NicelyFormattedTypes,
} from "@rilldata/web-local/lib/util/humanize-numbers";

import {
  splitNumStr,
  getSpacingMetadataForRawStrings,
  getSpacingMetadataForSplitStrings,
} from "./num-string-to-aligned-spec";
import type {
  FormatterFactory,
  FormatterSpacingMeta,
  NumberStringParts,
} from "./number-to-string-formatters";

const ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX = {
  0: "",
  3: "k",
  6: "M",
  9: "B",
  12: "T",
  15: "Q",
};

const shortScaleSuffixIfAvailable = (x: number): string => {
  let suffix = ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX[x];
  if (suffix !== undefined) return suffix;

  return "E" + x;
};

const formatNumWithOrderOfMag = (
  x: number,
  newOrder: number,
  options = { minimumFractionDigits: 3 }
): NumberStringParts => {
  const [int, frac] = Intl.NumberFormat("en-US", options)
    .format(x / 10 ** newOrder)
    .split(".");
  const splitStr = { int, frac, suffix: "E" + newOrder };

  return splitStr;
};

const thousandthsNumAsDecimalNumParts = (
  x: number,
  maximumFractionDigits: number = 6,
  padZero: boolean = false
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

  return { int, frac, suffix: "" };
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
    let formatter = new Intl.NumberFormat("en-US", {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    });

    splitStrs = sample
      .map((x) => formatter.format(x).replace(",", ""))
      .map(splitNumStr);
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
  let rawStrings = sample.map(engFmt.format);
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

    let suffix = shortScaleSuffixIfAvailable(ordersOfMag[i]);
    return { ...ss, ...{ suffix } };
  });

  return splitStrs;
};

export const humanized2FormatterFactory: FormatterFactory = (
  sample: number[],
  options
) => {
  const engFmt = new Intl.NumberFormat("en-US", {
    notation: "engineering",
    minimumFractionDigits: 3,
  });
  let rawStrings = sample.map(engFmt.format);
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  let ordersOfMag = splitStrs.map((ss) => +ss.suffix.slice(1));

  // omit exact zeros when calculating orders of magnitude
  const orderOfMagNoZeros = ordersOfMag.filter((_, i) => sample[i] != 0);
  let maxOrder = Math.max(...orderOfMagNoZeros);
  let minOrder = Math.min(...orderOfMagNoZeros);

  splitStrs.forEach((ss, i) => {
    let suff = ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX[ordersOfMag[i]];
    if (suff !== undefined) ss.suffix = suff;
  });

  splitStrs.forEach((ss) => {
    if (ss.suffix === undefined) console.log("bad suffix pre", ss);
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

    default:
      break;
  }

  splitStrs.forEach((ss, i) => {
    if (ss.suffix === undefined) console.log("bad suffix post", ss);
  });

  let spacing: FormatterSpacingMeta =
    getSpacingMetadataForSplitStrings(splitStrs);

  return (x: number) => {
    let i = sample.findIndex((h) => h === x);
    return {
      number: x,
      rawStr: rawStrings[i],
      splitStr: splitStrs[i],
      spacing,
    };
  };
};
