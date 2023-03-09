// Create types and then present an appropriate string
// Current dash persion has `prefix` key in JSON to add currecny etc.
// We can provide a dropdown option in the table?? or regex??

import { humanizedFormatterFactory } from "@rilldata/web-common/lib/number-formatting/humanizer";
import {
  FormatterFactoryOptions,
  NumberKind,
} from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import type { LeaderboardValue } from "./dashboard-stores";

const shortHandSymbols = ["Q", "T", "B", "M", "k", "none"] as const;
export type ShortHandSymbols = typeof shortHandSymbols[number];

interface HumanizeOptions {
  scale?: ShortHandSymbols;
  excludeDecimalZeros?: boolean;
  columnName?: string;
}

type formatterOptions = Intl.NumberFormatOptions & HumanizeOptions;

const shortHandMap = {
  Q: 1.0e15,
  T: 1.0e12,
  B: 1.0e9,
  M: 1.0e6,
  k: 1.0e3,
  none: 1,
};

export enum NicelyFormattedTypes {
  HUMANIZE = "humanize",
  NONE = "none",
  CURRENCY = "currency_usd",
  PERCENTAGE = "percentage",
  DECIMAL = "comma_separators",
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
  { value: NicelyFormattedTypes.DECIMAL, label: "Decimal" },
];

const DEFAULT_OPTIONS = {
  locale: "en-US",
  style: "decimal",
  currency: "USD",
  maximumFractionDigits: 2,
  currencyDisplay: "narrowSymbol",
  currencySign: "standard",
};

function getNumberFormatter(
  type: NicelyFormattedTypes,
  options?: formatterOptions
): Intl.NumberFormat {
  const o = { ...DEFAULT_OPTIONS, ...(options || {}) };

  if (type == NicelyFormattedTypes.CURRENCY) {
    o.style = "currency";
  } else if (type == NicelyFormattedTypes.PERCENTAGE) {
    o.style = "percent";
    o.maximumFractionDigits = 4;
  }
  const { locale, ...opts } = o;
  return new Intl.NumberFormat(locale, opts);
}

function formatNicely(
  value: number,
  type: NicelyFormattedTypes,
  options?: formatterOptions
): string {
  const formatterOptions = Object.assign({}, options);
  if (options?.excludeDecimalZeros) {
    delete formatterOptions["excludeDecimalZeros"];
  }

  const formatter = getNumberFormatter(type, formatterOptions);
  return formatter.format(value);
}

function convertToShorthand(value: number): string | number {
  if (Math.abs(value) < 1000)
    return formatNicely(value, NicelyFormattedTypes.DECIMAL);

  // Fifteen Zeros for Quadrillion
  return Math.abs(value) >= 1.0e15
    ? (value / 1.0e15).toFixed(1) + "Q"
    : // Twelve Zeros for Trillions
    Math.abs(value) >= 1.0e12
    ? (value / 1.0e12).toFixed(1) + "T"
    : // Nine Zeroes for Billions
    Math.abs(value) >= 1.0e9
    ? (value / 1.0e9).toFixed(1) + "B"
    : // Six Zeroes for Millions
    Math.abs(value) >= 1.0e6
    ? (value / 1.0e6).toFixed(1) + "M"
    : // Three Zeroes for Thousands
    Math.abs(value) >= 1.0e3
    ? (value / 1.0e3).toFixed(1) + "k"
    : value;
}

function getScaleForValue(value: number): ShortHandSymbols {
  return Math.abs(value) >= 1.0e15
    ? "Q"
    : Math.abs(value) >= 1.0e12
    ? "T"
    : Math.abs(value) >= 1.0e9
    ? "B"
    : Math.abs(value) >= 1.0e6
    ? "M"
    : Math.abs(value) >= 1.0e3
    ? "k"
    : "none";
}

/*
  Format a single value using the given type and options
*/
export function humanizeDataType_legacy(
  value: unknown,
  type: NicelyFormattedTypes,
  options?: formatterOptions
) {
  let formattedValue;
  if (typeof value != "number" || type == NicelyFormattedTypes.NONE)
    return value;
  else if (type == NicelyFormattedTypes.HUMANIZE) {
    formattedValue = convertToShorthand(value);
  } else if (type == NicelyFormattedTypes.CURRENCY) {
    formattedValue = "$" + convertToShorthand(value);
  } else {
    return formatNicely(value, type, options);
  }

  if (formattedValue && options?.excludeDecimalZeros) {
    return formattedValue.replace(".0", "");
  } else {
    return formattedValue;
  }
}

function determineScaleForValues(values: number[]): ShortHandSymbols {
  let numberValues = values;
  const nullIndex = values.indexOf(null);
  if (nullIndex !== -1) {
    numberValues = values.slice(0, nullIndex);
  }

  // Convert negative numbers to absolute
  numberValues = numberValues.map((v) => Math.abs(v)).sort((a, b) => b - a);

  const half = Math.floor(numberValues.length / 2);
  let median: number;
  if (numberValues.length % 2) median = numberValues[half];
  else median = (numberValues[half - 1] + numberValues[half]) / 2.0;

  let scaleForMax = getScaleForValue(numberValues[0]);
  while (scaleForMax != shortHandSymbols[shortHandSymbols.length - 1]) {
    const medianShorthand = (
      Math.abs(median) / shortHandMap[scaleForMax]
    ).toFixed(1);

    const numDigitsInMedian = medianShorthand.toString().split(".")[0].length;
    if (numDigitsInMedian >= 1) {
      return scaleForMax;
    } else {
      scaleForMax = shortHandSymbols[shortHandSymbols.indexOf(scaleForMax) + 1];
    }
  }
  return scaleForMax;
}

function applyScaleOnValues(values: number[], scale: ShortHandSymbols) {
  if (scale == shortHandSymbols[shortHandSymbols.length - 1]) {
    const formatter = getNumberFormatter(NicelyFormattedTypes.DECIMAL);
    return values.map((v) => {
      if (v === null) return "∅";
      else return formatter.format(v);
    });
  }
  return values.map((v) => {
    if (v === null) return "∅";
    const shortHandNumber = v / shortHandMap[scale];
    let shortHandValue: string;
    if (Math.abs(shortHandNumber) < 0.1) {
      shortHandValue = "<0.1";
    } else {
      shortHandValue = shortHandNumber.toFixed(1);
    }

    return shortHandValue + scale;
  });
}

function humanizeGroupValuesUtil(
  values: number[],
  type: NicelyFormattedTypes,
  options?: formatterOptions
) {
  if (!values.length) return values;
  if (type == NicelyFormattedTypes.NONE) return values;
  else if (type == NicelyFormattedTypes.HUMANIZE) {
    let scale;
    if (options?.scale) {
      scale = options.scale;
    } else scale = determineScaleForValues(values);
    return applyScaleOnValues(values, scale);
  } else if (type == NicelyFormattedTypes.CURRENCY) {
    let scale;
    if (options?.scale) {
      scale = options.scale;
    } else scale = determineScaleForValues(values);
    return applyScaleOnValues(values, scale).map((v) => "$" + v);
  } else {
    let formatterOptions = {};
    formatterOptions = Object.assign({}, options);
    delete formatterOptions["scale"];
    delete formatterOptions["columnName"];
    const formatter = getNumberFormatter(type, formatterOptions);
    return values.map((v) => {
      if (v === null) return "∅";
      else return formatter.format(v);
    });
  }
}

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
  const formattedValues = humanizeGroupValuesUtil2(
    numValues as number[],
    type,
    options
  );

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

export function getScaleForLeaderboard(
  leaderboard: Map<string, Array<LeaderboardValue>>
) {
  if (!leaderboard) return "none";

  const numValues = [...leaderboard.values()]
    // use the first five dimensions as the sample
    .slice(0, 5)
    // Take only first 7 values which are shown as input
    .map((values) => values.slice(0, 7))
    .flat()
    .map((values) => values.value);

  const areAllNumbers = numValues.every((e) => typeof e === "number");
  if (!areAllNumbers) return "none";

  const sortedValues = numValues.sort((a, b) => b - a);

  return determineScaleForValues(sortedValues);
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
      // NicelyFormattedTypes.DECIMAL
      // NicelyFormattedTypes.NONE
      // NicelyFormattedTypes.HUMANIZE
      return NumberKind.ANY;
  }
};

export function humanizeDataType(
  value: unknown,
  type: NicelyFormattedTypes
): string {
  if (typeof value != "number") return value.toString();

  const numberKind = nicelyFormattedTypesToNumberKind(type);

  let innerOptions: FormatterFactoryOptions;
  if (type === NicelyFormattedTypes.NONE) {
    innerOptions = {
      strategy: "none",
      numberKind,
      padWithInsignificantZeros: false,
    };
  } else {
    innerOptions = {
      strategy: "default",
      numberKind,
      padWithInsignificantZeros: false,
      maxDigitsRightSmallNums: 1,
      maxDigitsRightSuffixNums: 2,
    };
  }

  return humanizedFormatterFactory([value], innerOptions).stringFormat(value);
}

/** This function is used primarily in the leaderboard and the detail tables. */
function humanizeGroupValuesUtil2(
  values: number[],
  type: NicelyFormattedTypes,
  options?: formatterOptions
) {
  if (!values.length) return values;
  if (type == NicelyFormattedTypes.NONE) return values;

  const numberKind = nicelyFormattedTypesToNumberKind(type);

  const innerOptions: FormatterFactoryOptions = {
    strategy: "default",
    numberKind,
    padWithInsignificantZeros: false,
    maxDigitsRightSmallNums: 2,
    maxDigitsRightSuffixNums: 1,
  };

  const formatter = humanizedFormatterFactory(values, innerOptions);

  return values.map((v) => {
    if (v === null) return "∅";
    else return formatter.stringFormat(v);
  });
}
