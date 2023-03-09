import {
  Formatter,
  FormatterFactory,
  FormatterFactoryOptions,
  NumberKind,
} from "./humanizer-types";
import { DefaultHumanizer } from "./strategies/default";
import { NonFormatter } from "./strategies/none";

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
      formatter = new DefaultHumanizer(sample, options);
      break;

    default:
      console.warn(
        `Number formatter strategy "${options.strategy}" is not implemented, using default strategy`
      );

      const defaultOptions: FormatterFactoryOptions = {
        strategy: "default",
        padWithInsignificantZeros: true,
        numberKind: options.numberKind || NumberKind.ANY,
        maxDigitsRightSmallNums: 3,
        maxDigitsRightSuffixNums: 2,
      };

      formatter = new DefaultHumanizer(sample, defaultOptions);
      break;
  }

  return formatter;
};
