import { Formatter, FormatterFactory, NumberKind } from "./humanizer-types";
import { NonFormatter } from "./strategies/none";
import {
  defaultDollarOptions,
  defaultGenericNumOptions,
  defaultPercentOptions,
  PerRangeFormatter,
} from "./strategies/per-range";

export const humanizedFormatterFactory: FormatterFactory = (
  sample: number[],
  options
) => {
  let formatter: Formatter;
  switch (options.strategy) {
    case "none":
      formatter = new NonFormatter(sample, options);
      break;

    case "default":
      // delegate to the range strategy formatter with
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

    default:
      console.warn(
        `Number formatter strategy "${options.strategy}" is not implemented, using default strategy`
      );
      formatter = new PerRangeFormatter(sample, defaultGenericNumOptions);
      break;
  }

  return formatter;
};
