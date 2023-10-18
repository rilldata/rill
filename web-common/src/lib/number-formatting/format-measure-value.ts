import {
  FormatPreset,
  formatPresetToNumberKind,
} from "@rilldata/web-common/features/dashboards/humanize-numbers";
import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import { format as d3format } from "d3-format";
import type { FormatterFactoryOptions } from "./humanizer-types";
import {
  formatMsInterval,
  formatMsToDuckDbIntervalString,
} from "./strategies/intervals";
import { humanizedFormatterFactory } from "./humanizer";

function humanizeDataType(value: unknown, type: FormatPreset): string {
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
function humanizeDataTypeUnabridged(
  value: unknown,
  type: FormatPreset
): string {
  if (type === FormatPreset.INTERVAL) {
    return formatMsToDuckDbIntervalString(value as number);
  }
  return value.toString();
}

export const createMeasureValueFormatter = (
  measureSpec: MetricsViewSpecMeasureV2,
  useUnabridged = false
): ((value: number) => string) => {
  const humanizer = useUnabridged
    ? humanizeDataTypeUnabridged
    : humanizeDataType;

  // Humanize by default if measureSpec is not provided.
  // This may e.g. be the case during the initial render of a dashboard,
  // when a measureSpec has not yet loaded from a metadata query.
  if (measureSpec === undefined)
    return (value: number) => humanizer(value, FormatPreset.HUMANIZE);

  // Use the d3 formatter if it is provided and valid
  // (d3 throws an error if the format is invalid).
  // otherwise, use the humanize formatter.
  if (measureSpec.formatD3 !== undefined && measureSpec.formatD3 !== "") {
    try {
      return d3format(measureSpec.formatD3);
    } catch (error) {
      return (value: number) => humanizer(value, FormatPreset.HUMANIZE);
    }
  }

  // finally, use the formatPreset.
  return (value: number) =>
    humanizer(
      value,
      (measureSpec.formatPreset as FormatPreset) ?? FormatPreset.HUMANIZE
    );
};
