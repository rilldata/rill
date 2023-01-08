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

type FormatterSpacingMeta = {
  maxWholeDigits: number;
  maxFracDigits: number;
  // maxFracDigitsWithSuffix: number;
  maxSuffixChars: number;
};

type NumToRawStringFormatter = (x: number) => string;
type NumToHtmlStringFormatter = (x: number) => string;
type RawStrToHtmlStrFormatter = (s: string) => string;

export type NumberFormatter = (x: number) => RichFormatNumber;

export type NumberStringParts = {
  int: string;
  frac: string;
  suffix: string;
};

export type RichFormatNumber = {
  number: number;
  rawStr: string;
  splitStr: NumberStringParts;
  spacing: FormatterSpacingMeta;
};

type FormatterFactory = (sample: number[], options) => NumberFormatter;

let humanizeGroupValuesFormatterFactory: FormatterFactory = (
  sample: number[],
  _options
) => {
  let humanized = humanizeGroupValues(
    sample.map((x) => ({ value: x })),
    NicelyFormattedTypes.HUMANIZE,
    { columnName: "value" }
  );
  let rawStrings = humanized.map((x) => x.__formatted_value.toString());
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  let spacing: FormatterSpacingMeta =
    getSpacingMetadataForRawStrings(rawStrings);

  return (x: number) => {
    let i = humanized.findIndex((h) => h.value == x);
    return {
      number: x,
      rawStr: rawStrings[i],
      splitStr: splitStrs[i],
      spacing,
    };
  };
};

let rawStrFormatterFactory: FormatterFactory = (sample: number[], _options) => {
  let rawStrings = sample.map((x) => x.toString());
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  let spacing: FormatterSpacingMeta =
    getSpacingMetadataForRawStrings(rawStrings);

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

let IntlFormatterFactory: FormatterFactory = (sample: number[], _options) => {
  let intlFormatter = new Intl.NumberFormat("en-US", {
    notation: "scientific",
  });

  let rawStrings = sample.map((x) => intlFormatter.format(x));
  let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

  let spacing: FormatterSpacingMeta =
    getSpacingMetadataForRawStrings(rawStrings);

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

let IntlFormatterFactoryWithBaseOptions =
  (baseOptions) => (sample: number[], options) => {
    let intlFormatter = new Intl.NumberFormat("en-US", {
      ...baseOptions,
      ...options,
    });

    let rawStrings = sample.map((x) => intlFormatter.format(x));
    let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

    let spacing: FormatterSpacingMeta =
      getSpacingMetadataForRawStrings(rawStrings);

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

let formatterFactoryFromStringifier =
  (stringifier: (number) => string) => (sample: number[], options) => {
    let rawStrings = sample.map(stringifier);
    let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

    let spacing: FormatterSpacingMeta =
      getSpacingMetadataForRawStrings(rawStrings);

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

let formatterFactoryFromStringifierWithOptions =
  (stringifierWithOptions: (options) => (number) => string) =>
  (sample: number[], options) => {
    let rawStrings = sample.map(stringifierWithOptions(options));
    let splitStrs: NumberStringParts[] = rawStrings.map(splitNumStr);

    let spacing: FormatterSpacingMeta =
      getSpacingMetadataForRawStrings(rawStrings);

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

const ORDER_OF_MAG_TO_LONG_SCALE_SUFFIX = {
  0: "",
  3: "k",
  6: "M",
  9: "B",
  12: "T",
  15: "Q",
};

const longScaleSuffixIfAvailable = (x: number): string => {
  let suffix = ORDER_OF_MAG_TO_LONG_SCALE_SUFFIX[x];
  if (suffix !== undefined) return suffix;
  return "E" + x;
};

let humanized2FormatterFactory: FormatterFactory = (
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
  let maxOrder = Math.max(...ordersOfMag);
  let minOrder = Math.min(...ordersOfMag);
  // console.log({ ordersOfMag, maxOrder, minOrder });

  splitStrs.forEach((ss, i) => {
    let suff = ORDER_OF_MAG_TO_LONG_SCALE_SUFFIX[ordersOfMag[i]];
    if (suff !== undefined) ss.suffix = suff;
  });

  splitStrs.forEach((ss) => {
    if (ss.suffix === undefined) console.log("bad suffix pre", ss);
  });

  if (options.onlyUseLargestMagnitude === true) {
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
    } else if (options.usePlainNumForThousandths && maxOrder === -3) {
      const formatter = new Intl.NumberFormat("en-US", {
        minimumFractionDigits: options.usePlainNumForThousandthsPadZeros
          ? 6
          : 1,
        maximumFractionDigits: 6,
      });

      splitStrs = sample.map((x) => formatter.format(x)).map(splitNumStr);
    } else {
      let thousandthsTruncator = new Intl.NumberFormat("en-US", {
        minimumFractionDigits: 3,
      });

      splitStrs = splitStrs.map((ss, i) => {
        let mag = ordersOfMag[i];
        let num = sample[i];
        let maxOrderSuffix: string = longScaleSuffixIfAvailable(maxOrder);

        if (mag !== maxOrder) {
          let newNum = Math.abs(num / 10 ** maxOrder);
          let resplit = thousandthsTruncator.format(newNum).split(".");
          let frac = resplit.length == 2 ? resplit[1] : resplit[0];
          return {
            int: num >= 0 ? "0" : "-0",
            frac,
            suffix: maxOrderSuffix,
          };
        } else {
          return ss;
        }
      });
    }
  }

  splitStrs.forEach((ss) => {
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

  { desc: "humanized 2", fn: humanized2FormatterFactory },

  {
    desc: "humanized 2, truncate small magnitudes",
    fn: (sample, options) =>
      humanized2FormatterFactory(sample, {
        ...options,
        useLargestMagnitudeOnly: true,
      }),
  },

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
