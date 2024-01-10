import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import { format as d3format } from "d3-format";
import {
  FormatPreset,
  formatPresetToNumberKind,
  Formatter,
  NumberKind,
  type FormatterFactoryOptions,
} from "./humanizer-types";
import {
  formatMsInterval,
  formatMsToDuckDbIntervalString,
  IntervalFormatter,
} from "./strategies/intervals";
import { PerRangeFormatter } from "./strategies/per-range";
import {
  defaultDollarOptions,
  defaultGenericNumOptions,
  defaultPercentOptions,
} from "./strategies/per-range-default-options";
import { NonFormatter } from "./strategies/none";

/**
 * This function is intended to provides a compact,
 * potentially lossy, humanized string representation of a number.
 */
function humanizeDataType(value: number, preset: FormatPreset): string {
  if (typeof value !== "number") {
    console.warn(
      `humanizeDataType only accepts numbers, got ${value} for FormatPreset "${preset}"`,
    );

    return JSON.stringify(value);
  }
  // if (type === FormatPreset.INTERVAL) return formatMsInterval(value);

  const numberKind = formatPresetToNumberKind(preset);

  let options: FormatterFactoryOptions;

  if (preset === FormatPreset.NONE) {
    options = {
      numberKind,
      padWithInsignificantZeros: false,
    };
  } else {
    options = {
      numberKind,
    };
  }

  // return humanizedFormatterFactory([value], options).stringFormat(value);

  const sample = [];

  switch (preset) {
    case FormatPreset.NONE:
      return new NonFormatter(sample, options).stringFormat(value);

    case FormatPreset.CURRENCY_USD:
      return new PerRangeFormatter(sample, defaultDollarOptions).stringFormat(
        value,
      );

    case FormatPreset.CURRENCY_EUR:
      return new PerRangeFormatter(sample, defaultDollarOptions).stringFormat(
        value,
      );

    case FormatPreset.PERCENTAGE:
      return new PerRangeFormatter(sample, defaultPercentOptions).stringFormat(
        value,
      );

    case FormatPreset.INTERVAL:
      return formatMsInterval(value);

    case FormatPreset.HUMANIZE:
      return new PerRangeFormatter(
        sample,
        defaultGenericNumOptions,
      ).stringFormat(value);

    default:
      console.warn(
        "Unknown format preset, using default formatter. All number kinds should be handled.",
      );
      return new PerRangeFormatter(
        sample,
        defaultGenericNumOptions,
      ).stringFormat(value);
  }

  // switch (options.strategy) {
  //   case "none":
  //     formatter = new NonFormatter(sample, options);
  //     break;

  //   case "default":
  //     // default strategy simply
  //     // delegates to the range strategy formatter with
  //     // appropriate default presets for NumberKind
  //     switch (options.numberKind) {
  //       case NumberKind.DOLLAR:
  //         formatter = new PerRangeFormatter(sample, defaultDollarOptions);
  //         break;
  //       case NumberKind.EURO:
  //         formatter = new PerRangeFormatter(sample, defaultDollarOptions);
  //         break;
  //       case NumberKind.PERCENT:
  //         formatter = new PerRangeFormatter(sample, defaultPercentOptions);
  //         break;
  //       case NumberKind.INTERVAL:
  //         formatter = new IntervalFormatter();
  //         break;
  //       case NumberKind.ANY:
  //         formatter = new PerRangeFormatter(sample, defaultGenericNumOptions);
  //         break;
  //       default:
  //         console.warn(
  //           "Unknown number kind, using default formatter. All number kinds should be handled.",
  //         );
  //         formatter = new PerRangeFormatter(sample, defaultGenericNumOptions);
  //         break;
  //     }
  //     break;
  // }

  // return formatter.stringFormat(value);
}

/**
 * This function is intended to provide a lossless
 * humanized string representation of a number in cases
 * where a raw number will be meaningless to the user.
 */
function humanizeDataTypeUnabridged(value: number, type: FormatPreset): string {
  if (typeof value !== "number") {
    console.warn(
      `humanizeDataTypeUnabridged only accepts numbers, got ${value} for FormatPreset "${type}"`,
    );
    return JSON.stringify(value);
  }
  if (type === FormatPreset.INTERVAL) {
    return formatMsToDuckDbIntervalString(value as number);
  }
  return value.toString();
}

/**
 * This higher-order function takes a measure spec and returns
 * a function appropriate for formatting values from that measure.
 *
 * You may optionally add type paramaters to allow non-numeric null
 * undefined values to be passed through unmodified.
 * - `createMeasureValueFormatter<null | undefined>(measureSpec)` will pass through null and undefined values unchanged
 * - `createMeasureValueFormatter<null>(measureSpec)` will pass through null values unchanged
 * - `createMeasureValueFormatter<undefined>(measureSpec)` will pass through undefined values unchanged
 *
 *
 * FIXME: we want to remove the need for this to *ever* accept undefined values,
 * as we switch to always using `null` to represent missing values.
 */
export function createMeasureValueFormatter<T extends null | undefined = never>(
  measureSpec: MetricsViewSpecMeasureV2,
  useUnabridged = false,
): (value: number | T) => string | T {
  const humanizer = useUnabridged
    ? humanizeDataTypeUnabridged
    : humanizeDataType;

  // Return and empty string if measureSpec is not provided.
  // This may e.g. be the case during the initial render of a dashboard,
  // when a measureSpec has not yet loaded from a metadata query.
  if (measureSpec === undefined) {
    return (value: number | T) => (typeof value === "number" ? "" : value);
  }

  // Use the d3 formatter if it is provided and valid
  // (d3 throws an error if the format is invalid).
  // otherwise, use the humanize formatter.
  if (measureSpec.formatD3 !== undefined && measureSpec.formatD3 !== "") {
    try {
      const formatter = d3format(measureSpec.formatD3);
      return (value: number | T) =>
        typeof value === "number" ? formatter(value) : value;
    } catch (error) {
      return (value: number | T) =>
        typeof value === "number"
          ? humanizer(value, FormatPreset.HUMANIZE)
          : value;
    }
  }

  // finally, use the formatPreset.
  const formatPreset =
    measureSpec.formatPreset && measureSpec.formatPreset !== ""
      ? (measureSpec.formatPreset as FormatPreset)
      : FormatPreset.HUMANIZE;

  return (value: number | T) =>
    typeof value === "number" ? humanizer(value, formatPreset) : value;
}
