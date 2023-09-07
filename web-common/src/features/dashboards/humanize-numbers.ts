// Create types and then present an appropriate string
// Current dash persion has `prefix` key in JSON to add currecny etc.
// We can provide a dropdown option in the table?? or regex??

import { humanizedFormatterFactory } from "@rilldata/web-common/lib/number-formatting/humanizer";
import {
  FormatterFactoryOptions,
  NumberKind,
} from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import {
  formatMsInterval,
  formatMsToDuckDbIntervalString,
} from "@rilldata/web-common/lib/number-formatting/strategies/intervals";
import { PerRangeFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/per-range";

const shortHandSymbols = ["Q", "T", "B", "M", "k", "none"] as const;
export type ShortHandSymbols = (typeof shortHandSymbols)[number];

/**
 * This enum represents all of the valid strings that can be
 * used in the `format_preset` field of a measure definition.
 */
export enum FormatPreset {
  HUMANIZE = "humanize",
  NONE = "none",
  CURRENCY = "currency_usd",
  PERCENTAGE = "percentage",
  INTERVAL = "interval_ms",
}

interface ColFormatSpec {
  columnName: string;
  formatPreset: FormatPreset;
}

export function humanizeGroupValues(
  values: Array<Record<string, number | string>>,
  type: FormatPreset,
  columnName?: string
) {
  const valueKey = columnName ?? "value";
  let numValues = values.map((v) => v[valueKey]);

  const areAllNumbers = numValues.some((e) => typeof e === "number");
  if (!areAllNumbers) return values;

  numValues = (numValues as number[]).sort((a, b) => b - a);
  const formattedValues = humanizeGroupValuesUtil2(numValues as number[], type);

  const formattedValueKey = "__formatted_" + valueKey;
  const humanizedValues = values.map((v) => {
    const index = numValues.indexOf(v[valueKey]);
    return { ...v, [formattedValueKey]: formattedValues[index] };
  });

  return humanizedValues;
}

export function humanizeGroupByColumns(
  values: Array<Record<string, number | string>>,
  columnFormatSpec: ColFormatSpec[]
) {
  return columnFormatSpec.reduce((valuesObj, column) => {
    return humanizeGroupValues(
      valuesObj,
      column.formatPreset || FormatPreset.HUMANIZE,
      column.columnName
    );
  }, values);
}

// NOTE: the following are adapters that I think fit the API
// used by the existing humanizer, but I'm not sure of the
// exact details, nor am I totally confident about the options
// passed in at all the relevant call sites, so I've added
// thes adapters rather than just pave over the existing functions.
// This really needs to be reviewed by Dhiraj, at which point we
// can deprecate any left over code that is no longer needed.

export const nicelyFormattedTypesToNumberKind = (
  type: FormatPreset | string
) => {
  switch (type) {
    case FormatPreset.CURRENCY:
      return NumberKind.DOLLAR;

    case FormatPreset.PERCENTAGE:
      return NumberKind.PERCENT;

    case FormatPreset.INTERVAL:
      return NumberKind.INTERVAL;

    default:
      // captures:
      // FormatPreset.NONE
      // FormatPreset.HUMANIZE
      return NumberKind.ANY;
  }
};

export function humanizeDataType(
  value: unknown,
  type: FormatPreset,
  options?: FormatterFactoryOptions
): string {
  if (value === undefined || value === null) return "";
  if (typeof value !== "number") return value.toString();

  const numberKind = nicelyFormattedTypesToNumberKind(type);

  let innerOptions: FormatterFactoryOptions = options;
  if (type === FormatPreset.NONE) {
    innerOptions = {
      strategy: "none",
      numberKind,
      padWithInsignificantZeros: false,
    };
  } else if (type === FormatPreset.INTERVAL) {
    return formatMsInterval(value);
  } else if (options === undefined) {
    innerOptions = {
      strategy: "default",
      numberKind,
    };
  } else {
    innerOptions = {
      strategy: "default",
      ...options,
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

/** This function is used primarily in the leaderboard and the detail tables. */
function humanizeGroupValuesUtil2(values: number[], type: FormatPreset) {
  if (!values.length) return values;
  if (type == FormatPreset.NONE) return values;

  const numberKind = nicelyFormattedTypesToNumberKind(type);

  const innerOptions: FormatterFactoryOptions = {
    strategy: "default",
    numberKind,
  };

  const formatter = humanizedFormatterFactory(values, innerOptions);

  return values.map((v) => {
    if (v === null) return "∅";
    else return formatter.stringFormat(v);
  });
}

/** formatter for the comparison percentage differences */
export function formatMeasurePercentageDifference(
  value,
  method = "partsFormat"
) {
  if (Math.abs(value * 100) < 1 && value !== 0) {
    return method === "partsFormat"
      ? { percent: "%", neg: "", int: "<1" }
      : "<1%";
  } else if (value === 0) {
    return method === "partsFormat" ? { percent: "%", neg: "", int: 0 } : "0%";
  }
  const factory = new PerRangeFormatter([], {
    strategy: "perRange",
    rangeSpecs: [
      {
        minMag: -2,
        supMag: 3,
        maxDigitsRight: 1,
        baseMagnitude: 0,
        padWithInsignificantZeros: false,
      },
    ],
    defaultMaxDigitsRight: 0,
    numberKind: NumberKind.PERCENT,
  });

  return factory[method](value);
}
