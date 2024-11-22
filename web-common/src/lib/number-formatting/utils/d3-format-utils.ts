import type { MetricsViewSpecMeasureV2FormatD3Locale } from "@rilldata/web-common/runtime-client";
import type { FormatLocaleDefinition } from "d3-format";

export function isValidD3Locale(
  config: MetricsViewSpecMeasureV2FormatD3Locale | undefined,
): boolean {
  if (!config) return false;
  if (config.currency) {
    // currency is an array of 2 strings
    if (!Array.isArray(config.currency) || config.currency.length !== 2)
      return false;
    return true;
  }
  return false;
}

export function getLocaleFromConfig(
  config: FormatLocaleDefinition,
): FormatLocaleDefinition {
  const base: FormatLocaleDefinition = {
    currency: ["$", ""],
    thousands: ",",
    grouping: [3],
    decimal: ".",
  };

  return { ...base, ...config };
}

export function currencyHumanizer(
  currency: [string, string],
  humanizedValue: string,
): string {
  const [prefix, suffix] = currency;
  // Replace the "$" symbol with the appropriate currency prefix/suffix
  return `${prefix}${humanizedValue.replace(/\$/g, "")}${suffix}`;
}

/**
 * Parse the currency symbol from a d3 format string.
 * For d3 the currency symbol is always "$" in the format string
 */
export function includesCurrencySymbol(formatString: string): boolean {
  return formatString.includes("$");
}
