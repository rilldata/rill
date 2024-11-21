import type { MetricsViewSpecMeasureV2FormatD3Locale } from "@rilldata/web-common/runtime-client";

export function isValidD3Locale(
  config: MetricsViewSpecMeasureV2FormatD3Locale | undefined,
): boolean {
  if (!config) return false;
  if (config.thousands && config.grouping && config.currency) {
    // thousands is a string
    if (typeof config.thousands !== "string") return false;
    // grouping is an array of numbers
    if (!Array.isArray(config.grouping)) return false;
    // currency is an array of 2 strings
    if (!Array.isArray(config.currency) || config.currency.length !== 2)
      return false;

    return true;
  }
  return false;
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
