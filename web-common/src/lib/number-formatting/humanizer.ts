import {
  FormatterFactory,
  NumberFormatter,
  NumberStringParts,
  NumberKind,
  FormatterWidths,
} from "./humanizer-types";
import { humanizeDefaultStrategy } from "./strategies/default";
import { orderOfMagnitude } from "./utils/format-with-order-of-magnitude";
import { smallestPrecisionMagnitude } from "./utils/smallest-precision-magnitude";

export const humanizedFormatterFactory: FormatterFactory = (
  sample: number[],
  options
) => {
  const range = { max: Math.max(...sample), min: Math.min(...sample) };

  const largestMagnitude = Math.max(...sample.map(orderOfMagnitude));
  const mostPreciseDigitMagnitude = Math.min(
    ...sample.map(smallestPrecisionMagnitude)
  );

  let splitStrs: NumberStringParts[];

  switch (options.strategy) {
    case "default":
      splitStrs = humanizeDefaultStrategy(sample, options);
      break;

    default:
      break;
  }

  splitStrs.forEach((ss, i) => {
    if (options.numberKind === NumberKind.DOLLAR) {
      ss.dollar = "$";
    }
    if (options.numberKind === NumberKind.PERCENT) {
      ss.percent = "%";
    }
  });

  let maxPxWidthsInSample: FormatterWidths;
  let maxPxWidthPossible: FormatterWidths;
  if (typeof options.pxWidthLookupFn === "function") {
    const maxPxWidthsInSample = getMaxPxWidthsForSplitsStrings(
      splitStrs,
      options.pxWidthLookupFn
    );
  }
  const { maxCharWidthsInSample, maxCharWidthPossible } =
    getCharWidthsForSplitStrs(splitStrs);

  const formatter: NumberFormatter = (x: number) => {
    let i = sample.findIndex((h) => h === x);
    return {
      number: x,
      splitStr: splitStrs[i],
      maxPxWidthsInSample,
      maxPxWidthPossible,
      maxCharWidthsInSample,
      maxCharWidthPossible,
      range,
      largestMagnitude,
      mostPreciseDigitMagnitude,
    };
  };
  return { formatter, options };
};
