// Create types and then present an appropriate string
// Current dash persion has `prefix` key in JSON to add currecny etc.
// We can provide a dropdown option in the table?? or regex??

const shortHandSymbols = ["B", "M", "k", "none"] as const;
type ShortHandSymbols = typeof shortHandSymbols[number];

const shortHandMap = {
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

export const nicelyFormattedTypesSelectorOptions = [
  { value: NicelyFormattedTypes.HUMANIZE, label: "Humanize - 12k" },
  { value: NicelyFormattedTypes.NONE, label: "No formatting - 12345.6789" },
  {
    value: NicelyFormattedTypes.CURRENCY,
    label: "Currency (USD) - $12,345.67",
  },
  { value: NicelyFormattedTypes.PERCENTAGE, label: "Percentage - 12345.6789%" },
  { value: NicelyFormattedTypes.DECIMAL, label: "Decimal - 12,345.67" },
];

console.log(
  "nicelyFormattedTypesSelectorOptions",
  nicelyFormattedTypesSelectorOptions
);

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
  options?: { [key: string]: any }
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
  options?: { [key: string]: any }
): string {
  const formatter = getNumberFormatter(type, options);
  return formatter.format(value);
}

function convertToShorthand(value: number): string | number {
  // Nine Zeroes for Billions
  return Math.abs(value) >= 1.0e9
    ? (Math.abs(value) / 1.0e9).toFixed(1) + "B"
    : // Six Zeroes for Millions
    Math.abs(value) >= 1.0e6
    ? (Math.abs(value) / 1.0e6).toFixed(1) + "M"
    : // Three Zeroes for Thousands
    Math.abs(value) >= 1.0e3
    ? (Math.abs(value) / 1.0e3).toFixed(1) + "k"
    : Math.abs(value);
}

function getScaleForValue(value: number): ShortHandSymbols {
  return Math.abs(value) >= 1.0e9
    ? "B"
    : Math.abs(value) >= 1.0e6
    ? "M"
    : Math.abs(value) >= 1.0e3
    ? "k"
    : "none";
}

export function humanizeDataType(
  value: number,
  type: NicelyFormattedTypes,
  options?: { [key: string]: any }
) {
  if (type == NicelyFormattedTypes.NONE) return value;
  else if (type == NicelyFormattedTypes.HUMANIZE) {
    if (value < 1000)
      return formatNicely(value, NicelyFormattedTypes.DECIMAL, options);
    else return convertToShorthand(value);
  } else {
    return formatNicely(value, type, options);
  }
}

function determineScaleForValues(values: number[]): ShortHandSymbols {
  const half = Math.floor(values.length / 2);
  let median: number;
  if (values.length % 2) median = values[half];
  else median = (values[half - 1] + values[half]) / 2.0;

  let scaleForMax = getScaleForValue(values[0]);

  while (scaleForMax != shortHandSymbols[shortHandSymbols.length - 1]) {
    const medianShorthand = (
      Math.abs(median) / shortHandMap[scaleForMax]
    ).toFixed(1);

    const numDigitsInMedian = medianShorthand.toString().split(".")[0].length;
    if (numDigitsInMedian >= 2) {
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
    return values.map((v) => formatter.format(v));
  }
  return values.map((v) => {
    const shortHandNumber = Math.abs(v) / shortHandMap[scale];
    let shortHandValue: string;
    if (shortHandNumber < 0.1) {
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
  options?: { [key: string]: any }
) {
  if (!values.length) return values;
  if (type == NicelyFormattedTypes.NONE) return values;
  else if (type == NicelyFormattedTypes.HUMANIZE) {
    const scale = determineScaleForValues(values);
    return applyScaleOnValues(values, scale);
  } else {
    const formatter = getNumberFormatter(type, options);
    return values.map((v) => formatter.format(v));
  }
}

export function humanizeGroupValues(
  values: Array<any>,
  type: NicelyFormattedTypes,
  options?: { [key: string]: any }
) {
  let numValues = values.map((v) => v.value);
  const areAllNumbers = numValues.every((e) => typeof e === "number");
  if (!areAllNumbers) return;

  numValues = (numValues as number[]).sort((a, b) => b - a);
  const formattedValues = humanizeGroupValuesUtil(
    numValues as number[],
    type,
    options
  );

  const humanizedValues = values.map((v) => {
    const index = numValues.indexOf(v.value);
    return { ...v, formattedValue: formattedValues[index] };
  });

  return humanizedValues;
}
