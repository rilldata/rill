import { Formatter, FormatterFactory, NumberKind } from "./humanizer-types";
import { IntTimesPowerOfTenFormatter } from "./strategies/IntTimesPowerOfTen";
import { NonFormatter } from "./strategies/none";
import { PerRangeFormatter } from "./strategies/per-range";
import {
  defaultDollarOptions,
  defaultGenericNumOptions,
  defaultPercentOptions,
} from "./strategies/per-range-default-options";

/**
 * This FormatterFactory is intended to be the user-facing
 * wrapper for formatters. This fumction delegates to a number
 * of different formatters depending upon the options
 * used, but by going through this wrapper those details
 * can be somewhat abstracted away in favor of config
 * options.
 *
 * @param sample
 * @param options
 * @returns Formatter
 */
export const humanizedFormatterFactory: FormatterFactory = (
  sample: number[],
  options
): Formatter => {
  let formatter: Formatter;

  switch (options.strategy) {
    case "none":
      formatter = new NonFormatter(sample, options);
      break;

    case "default":
      // default strategy simply
      // delegates to the range strategy formatter with
      // appropriate default presets for NumberKind
      switch (options.numberKind) {
        case NumberKind.DOLLAR:
          formatter = new PerRangeFormatter(sample, defaultDollarOptions);
          break;
        case NumberKind.PERCENT:
          formatter = new PerRangeFormatter(sample, defaultPercentOptions);
          break;
        default:
          formatter = new PerRangeFormatter(sample, defaultGenericNumOptions);
          break;
      }
      break;

    case "intTimesPowerOfTen":
      formatter = new IntTimesPowerOfTenFormatter(sample, options);
      break;

    default:
      console.warn(
        `Number formatter strategy "${options.strategy}" is not implemented, using default strategy`
      );
      formatter = new PerRangeFormatter(sample, defaultGenericNumOptions);
      break;
  }

  return formatter;
};

const percentHumanizer = humanizedFormatterFactory([], {
  strategy: "default",
  numberKind: NumberKind.PERCENT,
});
const countHumanizer = humanizedFormatterFactory([], {
  strategy: "default",
  numberKind: NumberKind.ANY,
});
const dollarHumanizer = humanizedFormatterFactory([], {
  strategy: "default",
  numberKind: NumberKind.DOLLAR,
});

// Re-exporting the default versions of the most common formatters
// in functional form. The extra machinery turns out to have been
// unneeded because we have decided not to use the stateful aspects
// that we thought would be required to format a sample of numbers
// in context.
export const humanizePercent = (value: number) =>
  percentHumanizer.stringFormat(value);
export const humanizeCount = (value: number) =>
  countHumanizer.stringFormat(value);
export const humanizeDollar = (value: number) =>
  dollarHumanizer.stringFormat(value);
