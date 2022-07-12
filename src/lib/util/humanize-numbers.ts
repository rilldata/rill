// Create types and then present an appropriate string
// Current dash persion has `prefix` key in JSON to add currecny etc.
// We can provide a dropdown option in the table?? or regex??

export enum NicelyFormattedTypes {
  CURRENCY = "currency",
  PERCENTAGE = "percentage",
  COUNT = "count",
  ACCOUNTING_CURRENCY = "accounting_currency",
  DECIMAL = "decimal",
  CURRENCY_WITH_NAME = "currency_with_name",
}

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
  } else if (type == NicelyFormattedTypes.ACCOUNTING_CURRENCY) {
    o.currencySign = "accounting";
    o.style = "currency";
  } else if (type == NicelyFormattedTypes.CURRENCY_WITH_NAME) {
    o.style = "currency";
    o.currencyDisplay = "name";
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
  return Math.abs(Number(value)) >= 1.0e9
    ? (Math.abs(Number(value)) / 1.0e9).toFixed(2) + "B"
    : // Six Zeroes for Millions
    Math.abs(Number(value)) >= 1.0e6
    ? (Math.abs(Number(value)) / 1.0e6).toFixed(2) + "M"
    : // Three Zeroes for Thousands
    Math.abs(Number(value)) >= 1.0e3
    ? (Math.abs(Number(value)) / 1.0e3).toFixed(2) + "k"
    : Math.abs(Number(value));
}

export function humanizeDataType(
  value: number,
  type: NicelyFormattedTypes,
  options?: { [key: string]: any }
) {
  if (type == NicelyFormattedTypes.COUNT) {
    return convertToShorthand(value);
  } else {
    return formatNicely(value, type, options);
  }
}
