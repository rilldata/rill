// Create types and then present an appropriate string
// Current dash persion has `prefix` key in JSON to add currecny etc.
// We can provide a dropdown option in the table?? or regex??

import { humanizedFormatterFactory } from "@rilldata/web-common/lib/number-formatting/humanizer";
import {
  FormatterFactoryOptions,
  NumberKind,
} from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { PerRangeFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/per-range";

const shortHandSymbols = ["Q", "T", "B", "M", "k", "none"] as const;
export type ShortHandSymbols = (typeof shortHandSymbols)[number];

interface HumanizeOptions {
  scale?: ShortHandSymbols;
  excludeDecimalZeros?: boolean;
  columnName?: string;
}

type formatterOptions = Intl.NumberFormatOptions & HumanizeOptions;

export enum NicelyFormattedTypes {
  HUMANIZE = "humanize",
  NONE = "none",
  CURRENCY = "currency_usd",
  PERCENTAGE = "percentage",
}

interface ColFormatSpec {
  columnName: string;
  formatPreset: NicelyFormattedTypes;
}

export const nicelyFormattedTypesSelectorOptions = [
  { value: NicelyFormattedTypes.HUMANIZE, label: "Humanize" },
  {
    value: NicelyFormattedTypes.NONE,
    label: "No formatting",
  },
  {
    value: NicelyFormattedTypes.CURRENCY,
    label: "Currency (USD)",
  },
  {
    value: NicelyFormattedTypes.PERCENTAGE,
    label: "Percentage",
  },
];

export function humanizeGroupValues(
  values: Array<Record<string, number | string>>,
  type: NicelyFormattedTypes,
  options?: formatterOptions
) {
  const valueKey = options.columnName ? options.columnName : "value";
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
      column.formatPreset || NicelyFormattedTypes.HUMANIZE,
      {
        columnName: column.columnName,
      }
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
  type: NicelyFormattedTypes | string
) => {
  switch (type) {
    case NicelyFormattedTypes.CURRENCY:
      return NumberKind.DOLLAR;

    case NicelyFormattedTypes.PERCENTAGE:
      return NumberKind.PERCENT;

    default:
      // captures:
      // NicelyFormattedTypes.NONE
      // NicelyFormattedTypes.HUMANIZE
      return NumberKind.ANY;
  }
};

export function humanizeDataType(
  value: unknown,
  type: NicelyFormattedTypes,
  options?: FormatterFactoryOptions
): string {
  if (value === undefined || value === null) return "";
  if (typeof value != "number") return value.toString();

  const numberKind = nicelyFormattedTypesToNumberKind(type);

  let innerOptions: FormatterFactoryOptions = options;
  if (type === NicelyFormattedTypes.NONE) {
    innerOptions = {
      strategy: "none",
      numberKind,
      padWithInsignificantZeros: false,
    };
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

/** This function is used primarily in the leaderboard and the detail tables. */
function humanizeGroupValuesUtil2(
  values: number[],
  type: NicelyFormattedTypes
) {
  if (!values.length) return values;
  if (type == NicelyFormattedTypes.NONE) return values;

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
