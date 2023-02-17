import {
  humanizeDataType,
  humanizeGroupValues,
  NicelyFormattedTypes,
} from "@rilldata/web-local/lib/util/humanize-numbers";
import { humanized2FormatterFactory } from "./humanizer-2";

import {
  splitNumStr,
  getSpacingMetadataForRawStrings,
  getSpacingMetadataForSplitStrings,
  getMaxPxWidthsForSplitsStrings,
} from "./num-string-to-aligned-spec";

export type FormatterSpacingMeta = {
  maxWholeDigits: number;
  maxFracDigits: number;
  // maxFracDigitsWithSuffix: number;
  maxSuffixChars: number;
};

export type NumToRawStringFormatter = (x: number) => string;
export type NumToHtmlStringFormatter = (x: number) => string;
export type RawStrToHtmlStrFormatter = (s: string) => string;

export type NumberFormatter = (x: number) => RichFormatNumber;

export type NumberStringParts = {
  int: string;
  dot: "" | ".";
  frac: string;
  suffix: string;
};

export type NumericRange = {
  min: number;
  max: number;
};

export type FormatterPxWidths = {
  int: number;
  dot: number;
  frac: number;
  suffix: number;
};

export type RichFormatNumber = {
  number: number;
  rawStr: string;
  splitStr: NumberStringParts;
  spacing: FormatterSpacingMeta;
  range: NumericRange;
  maxPxWidth: FormatterPxWidths;
};

export type FormatterFactory = (
  sample: number[],
  pxWidthLookup: NumPartPxWidthLookupFn,
  options
) => NumberFormatter;

export type NumPartPxWidthLookupFn = (str: string, isNumStr: boolean) => number;

let humanizeGroupValuesFormatterFactory: FormatterFactory = (
  sample: number[],
  pxWidthLookup: NumPartPxWidthLookupFn,
  _options
) => {
  let range = { max: Math.max(...sample), min: Math.min(...sample) };

  let humanized = humanizeGroupValues(
    sample.map((x) => ({ value: x })),
    NicelyFormattedTypes.HUMANIZE,
    { columnName: "value" }
  );
  let rawStrings = humanized.map((x) => x.__formatted_value.toString());
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  let spacing: FormatterSpacingMeta =
    getSpacingMetadataForRawStrings(rawStrings);

  const maxPxWidth = getMaxPxWidthsForSplitsStrings(splitStrs, pxWidthLookup);

  return (x: number) => {
    let i = humanized.findIndex((h) => h.value == x);
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

let rawStrFormatterFactory: FormatterFactory = (
  sample: number[],
  pxWidthLookup: NumPartPxWidthLookupFn,
  _options
) => {
  let range = { max: Math.max(...sample), min: Math.min(...sample) };

  let rawStrings = sample.map((x) => x.toString());
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  let spacing: FormatterSpacingMeta =
    getSpacingMetadataForRawStrings(rawStrings);
  const maxPxWidth = getMaxPxWidthsForSplitsStrings(splitStrs, pxWidthLookup);

  return (x: number) => {
    let i = sample.findIndex((h) => h === x);
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

// let IntlFormatterFactory: FormatterFactory = (sample: number[], _options) => {
//   let intlFormatter = new Intl.NumberFormat("en-US", {
//     notation: "scientific",
//   });

//   let rawStrings = sample.map((x) => intlFormatter.format(x));
//   let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

//   let spacing: FormatterSpacingMeta =
//     getSpacingMetadataForRawStrings(rawStrings);

//   return (x: number) => {
//     let i = sample.findIndex((h) => h === x);
//     return {
//       number: x,
//       rawStr: rawStrings[i],
//       splitStr: splitStrs[i],
//       spacing,
//     };
//   };
// };

let IntlFormatterFactoryWithBaseOptions =
  (baseOptions) =>
  (sample: number[], pxWidthLookup: NumPartPxWidthLookupFn, options) => {
    let range = { max: Math.max(...sample), min: Math.min(...sample) };

    let intlFormatter = new Intl.NumberFormat("en-US", {
      ...baseOptions,
      ...options,
    });

    let rawStrings = sample.map((x) => intlFormatter.format(x));
    let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

    let spacing: FormatterSpacingMeta =
      getSpacingMetadataForRawStrings(rawStrings);
    const maxPxWidth = getMaxPxWidthsForSplitsStrings(splitStrs, pxWidthLookup);

    return (x: number) => {
      let i = sample.findIndex((h) => h === x);
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

let formatterFactoryFromStringifier =
  (stringifier: (number) => string) =>
  (sample: number[], pxWidthLookup: NumPartPxWidthLookupFn, options) => {
    let range = { max: Math.max(...sample), min: Math.min(...sample) };

    let rawStrings = sample.map(stringifier);
    let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

    let spacing: FormatterSpacingMeta =
      getSpacingMetadataForRawStrings(rawStrings);
    const maxPxWidth = getMaxPxWidthsForSplitsStrings(splitStrs, pxWidthLookup);

    return (x: number) => {
      let i = sample.findIndex((h) => h === x);
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

let formatterFactoryFromStringifierWithOptions =
  (stringifierWithOptions: (options) => (number) => string) =>
  (sample: number[], pxWidthLookup: NumPartPxWidthLookupFn, options) => {
    let range = { max: Math.max(...sample), min: Math.min(...sample) };

    let rawStrings = sample.map(stringifierWithOptions(options));
    let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

    let spacing: FormatterSpacingMeta =
      getSpacingMetadataForRawStrings(rawStrings);

    const maxPxWidth = getMaxPxWidthsForSplitsStrings(splitStrs, pxWidthLookup);

    return (x: number) => {
      let i = sample.findIndex((h) => h === x);
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

type NamedFormatterFactory = {
  desc: string;
  fn: FormatterFactory;
};

export const formatterFactories: NamedFormatterFactory[] = [
  { desc: "JS `toString()`", fn: rawStrFormatterFactory },

  {
    desc: "humanizeGroupValues (current humanizer)",
    fn: humanizeGroupValuesFormatterFactory,
  },

  { desc: "new humanizer", fn: humanized2FormatterFactory },

  {
    desc: "scientific",
    fn: IntlFormatterFactoryWithBaseOptions({
      notation: "scientific",
    }),
  },
  // 9.877E8

  {
    desc: "engineering",
    fn: IntlFormatterFactoryWithBaseOptions({
      notation: "engineering",
    }),
  },
  // 987.654E6
  {
    desc: "compactShort + eng for small",
    fn: formatterFactoryFromStringifierWithOptions((options) => (x) => {
      if (Math.abs(x) < 0.01) {
        return new Intl.NumberFormat("en-US", {
          notation: "engineering",
          ...options,
        }).format(x);
      } else {
        return new Intl.NumberFormat("en-US", {
          notation: "compact",
          compactDisplay: "short",
          ...options,
        }).format(x);
      }
    }),
  },

  {
    desc: "compact, short",
    fn: IntlFormatterFactoryWithBaseOptions({
      notation: "compact",
      compactDisplay: "short",
    }),
  },

  {
    desc: "compact, long",
    fn: IntlFormatterFactoryWithBaseOptions({
      notation: "compact",
      compactDisplay: "long",
    }),
  },
  // 988 millions
];
