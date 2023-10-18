import {
  Formatter,
  FormatterFactory,
  FormatterFactoryOptions,
  NumberKind,
} from "./humanizer-types";
import { IntTimesPowerOfTenFormatter } from "./strategies/IntTimesPowerOfTen";
import { NonFormatter } from "./strategies/none";
import {
  IntervalFormatter,
  formatMsInterval,
  formatMsToDuckDbIntervalString,
} from "./strategies/intervals";
import { PerRangeFormatter } from "./strategies/per-range";
import {
  defaultDollarOptions,
  defaultGenericNumOptions,
  defaultPercentOptions,
} from "./strategies/per-range-default-options";
import {
  FormatPreset,
  formatPresetToNumberKind,
} from "@rilldata/web-common/features/dashboards/humanize-numbers";

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
        case NumberKind.INTERVAL:
          formatter = new IntervalFormatter();
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

export function humanizeDataType(value: unknown, type: FormatPreset): string {
  if (value === undefined || value === null) return "";
  if (typeof value !== "number") return value.toString();

  const numberKind = formatPresetToNumberKind(type);

  let innerOptions: FormatterFactoryOptions;

  if (type === FormatPreset.NONE) {
    innerOptions = {
      strategy: "none",
      numberKind,
      padWithInsignificantZeros: false,
    };
  } else if (type === FormatPreset.INTERVAL) {
    return formatMsInterval(value);
  } else {
    innerOptions = {
      strategy: "default",
      numberKind,
    };
  }
  return humanizedFormatterFactory([value], innerOptions).stringFormat(value);
}

/**
 * This function is intended to provide a lossless
 * humanized string representation of a number in cases
 * where a raw number will be meaningless to the user.
 */
export function humanizeDataTypeExpanded(
  value: unknown,
  type: FormatPreset
): string {
  if (type === FormatPreset.INTERVAL) {
    return formatMsToDuckDbIntervalString(value as number);
  }
  return value.toString();
}
