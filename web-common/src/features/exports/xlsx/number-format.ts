import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

export const PERCENT_FORMAT = "0.0%";
const INTEGER_FORMAT = "#,##0";
const DECIMAL_FORMAT = "#,##0.00";

// The Excel number format for a measure's cells. Presets map directly; d3
// format strings are translated where Excel has an equivalent; anything else
// falls back to a grouped integer or two-decimal format depending on the values.
export function excelNumberFormat(
  measure: MetricsViewSpecMeasure | undefined,
  values?: Iterable<unknown>,
): string {
  switch (measure?.formatPreset) {
    case FormatPreset.CURRENCY_USD:
      return `"$"#,##0.00`;
    case FormatPreset.CURRENCY_EUR:
      return `"€"#,##0.00`;
    case FormatPreset.PERCENTAGE:
      return percentFormat(values);
    case FormatPreset.INTERVAL:
    case FormatPreset.NONE:
      return "General";
  }
  const fromD3 = measure?.formatD3
    ? d3ToExcelFormat(
        measure.formatD3,
        values,
        localeCurrency(measure.formatD3Locale),
      )
    : undefined;
  return fromD3 ?? (allIntegers(values) ? INTEGER_FORMAT : DECIMAL_FORMAT);
}

// A percentage format with enough decimals to show the smallest non-zero value
// with two significant digits, as the dashboard's humanized percentages do
// (0.53%, 6.6%), keeping between one and four decimals. A fixed "0.0%" would
// turn a 0.53% click-through rate into 0.5% and a 0.03% one into 0.0%.
function percentFormat(values: Iterable<unknown> | undefined): string {
  let smallest = Infinity;
  for (const value of values ?? []) {
    if (typeof value === "number" && Number.isFinite(value) && value !== 0) {
      smallest = Math.min(smallest, Math.abs(value) * 100);
    }
  }
  const digits = Number.isFinite(smallest)
    ? 1 - Math.floor(Math.log10(smallest))
    : 1;
  return "0." + "0".repeat(Math.min(4, Math.max(1, digits))) + "%";
}

// The currency symbols of a measure's d3 locale, [prefix, suffix], when it sets them.
function localeCurrency(
  locale: MetricsViewSpecMeasure["formatD3Locale"],
): [string, string] | undefined {
  const currency = locale?.currency;
  if (
    Array.isArray(currency) &&
    currency.length === 2 &&
    currency.every((part) => typeof part === "string")
  ) {
    return [currency[0], currency[1]];
  }
  return undefined;
}

// d3-format specifier: [[fill]align][sign][symbol][0][width][,][.precision][~][type]
const D3_SPECIFIER =
  /^(?:(.)?([<>=^]))?([+\-( ])?([$#])?(0)?(\d+)?(,)?(\.\d+)?(~)?([a-z%])?$/i;

// `currency` is the [prefix, suffix] a "$" in the specifier stands for; d3's default is a dollar prefix.
export function d3ToExcelFormat(
  spec: string,
  values?: Iterable<unknown>,
  currency: [string, string] = ["$", ""],
): string | undefined {
  const match = D3_SPECIFIER.exec(spec.trim());
  if (!match) return undefined;
  const [, , , , symbol, , , comma, precision, , type] = match;
  const digits = precision ? Number(precision.slice(1)) : undefined;
  const grouped = comma ? "#,##0" : "0";
  const decimals = (count: number) =>
    count > 0 ? "." + "0".repeat(count) : "";

  let format: string;
  switch (type?.toLowerCase()) {
    case "%":
      format = "0" + decimals(digits ?? 0) + "%";
      break;
    case "d":
      format = grouped;
      break;
    case "f":
      format = grouped + decimals(digits ?? 2);
      break;
    case "s":
    case "e":
    case "g":
    case "r":
    case undefined:
      // SI prefixes and significant digits have no Excel equivalent; keep the
      // grouping and show decimals only when the values need them.
      format = grouped + decimals(digits ?? (allIntegers(values) ? 0 : 2));
      break;
    default:
      return undefined;
  }
  if (symbol !== "$") return format;
  const [prefix, suffix] = currency;
  return `${quoted(prefix)}${format}${quoted(suffix)}`;
}

// A literal for an Excel number format, or nothing for an empty string.
function quoted(text: string): string {
  return text ? `"${text.replace(/"/g, "")}"` : "";
}

function allIntegers(values: Iterable<unknown> | undefined): boolean {
  if (!values) return true;
  for (const value of values) {
    if (
      typeof value === "number" &&
      Number.isFinite(value) &&
      !Number.isInteger(value)
    ) {
      return false;
    }
  }
  return true;
}
