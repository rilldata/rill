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
