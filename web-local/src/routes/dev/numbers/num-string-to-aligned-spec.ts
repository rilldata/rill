// import type { NumToRawStringFnFactory } from "./number-to-string-formatters";

import type {
  FormatterMaxPxWidths,
  NumberStringParts,
  NumPartPxWidthLookupFn,
} from "./number-to-string-formatters";

export function splitNumStr(numStr: string): NumberStringParts {
  let nonNumReMatch = numStr.match(/[a-zA-z ]/);
  let int = "";
  let dot: "" | "." = numStr.includes(".") ? "." : "";
  let frac = "";
  let suffix = "";
  if (nonNumReMatch) {
    let suffixIndex = nonNumReMatch.index;
    let numPart = numStr.slice(0, suffixIndex);
    suffix = numStr.slice(suffixIndex);

    if (numPart.split(".").length == 1) {
      int = numPart;
    } else {
      int = numPart.split(".")[0];
      frac = numPart.split(".")[1] ?? "";
    }
  } else {
    int = numStr.split(".")[0];
    frac = numStr.split(".")[1] ?? "";
  }
  if (suffix === undefined) {
    console.log({ numStr, int, frac, suffix });
  }
  return { int, dot, frac, suffix };
}

// type AlignedNumberSpec = {
//   whole: string;
//   frac: string;
//   suffix: string;

//   wholeChars: number;
//   fracChars: number;
//   suffixChars: number;
// };

export const getSpacingMetadataForSplitStrings = (
  numStrParts: NumberStringParts[]
) => {
  return numStrParts
    .map((s) => {
      try {
        return {
          maxWholeDigits: s.int.length,
          maxFracDigits: s.frac.length,
          // maxFracDigitsWithSuffix: s.frac.length + s.suffix.length,
          maxSuffixChars: s?.suffix?.length ?? 0,
        };
      } catch (error) {
        console.log(s);
      }
    })
    .reduce(
      (a, b) => ({
        maxWholeDigits: Math.max(a.maxWholeDigits, b.maxWholeDigits),
        maxFracDigits: Math.max(a.maxFracDigits, b.maxFracDigits),
        // maxFracDigitsWithSuffix: Math.max(
        //   a.maxFracDigitsWithSuffix,
        //   b.maxFracDigitsWithSuffix
        // ),
        maxSuffixChars: Math.max(a.maxSuffixChars, b.maxSuffixChars),
      }),
      {
        maxWholeDigits: 0,
        maxFracDigits: 0,
        // maxFracDigitsWithSuffix: 0,
        maxSuffixChars: 0,
      }
    );
};

export const getSpacingMetadataForRawStrings = (numericStrings: string[]) => {
  return getSpacingMetadataForSplitStrings(numericStrings.map(splitNumStr));
};

export const getMaxPxWidthsForSplitsStrings = (
  numStrParts: NumberStringParts[],
  pxWidthLookup: NumPartPxWidthLookupFn
): FormatterMaxPxWidths => {
  let maxPxWidths = { int: 0, dot: 0, frac: 0, suffix: 0 };
  const max = Math.max;
  numStrParts.forEach((richNum) => {
    maxPxWidths.int = max(pxWidthLookup(richNum.int, true), maxPxWidths.int);
    maxPxWidths.dot = max(pxWidthLookup(richNum.dot, false), maxPxWidths.dot);
    maxPxWidths.frac = max(pxWidthLookup(richNum.frac, true), maxPxWidths.frac);
    maxPxWidths.suffix = max(
      pxWidthLookup(richNum.suffix, false),
      maxPxWidths.suffix
    );
  });
  return maxPxWidths;
};

// export const numStrToAlignedNumSpec = (
//   numToStrFactory: NumToRawStringFnFactory
// ) => {
//   return (sample: number[]) => {
//     const numToStr = numToStrFactory(sample);
//     let rawStrings = sample.map(numToStr);
//     let spacingMeta = getSpacingMetadataForStrings(rawStrings);

//     return (x: number) => {
//       let splitStr = splitNumStr(numToStr(x).toString());

//       return {
//         whole: splitStr.int,
//         frac: splitStr.frac,
//         suffix: splitStr.suffix,

//         wholeChars: spacingMeta.maxWholeDigits,
//         fracChars: spacingMeta.maxFracDigits,
//         suffixChars: spacingMeta.maxSuffixChars,
//       };
//     };
//   };
// };
