import {
  getLocaleFromConfig,
  includesCurrencySymbol,
  isValidD3Locale,
} from "@rilldata/web-common/lib/number-formatting/utils/d3-format-utils";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import {
  format as d3format,
  formatLocale as d3FormatLocale,
  type FormatLocaleDefinition,
} from "d3-format";
import memoize from "memoizee";
import {
  FormatPreset,
  type LocaleConfig,
  NumberKind,
  type ContextOptions,
  type FormatterContext,
  type FormatterContextSurface,
} from "./humanizer-types";
import {
  formatMsInterval,
  formatMsToDuckDbIntervalString,
} from "./strategies/intervals";
import { PerRangeFormatter } from "./strategies/per-range";
import {
  axisCurrencyOptions,
  axisDefaultFormattingOptions,
  axisPercentOptions,
} from "./strategies/per-range-axis-options";
import {
  bigNumCurrencyOptions,
  bigNumDefaultFormattingOptions,
  bigNumPercentOptions,
} from "./strategies/per-range-bignum-options";
import {
  defaultCurrencyOptions,
  defaultGenericNumOptions,
  defaultNoFormattingOptions,
  defaultPercentOptions,
} from "./strategies/per-range-default-options";
import {
  tooltipCurrencyOptions,
  tooltipNoFormattingOptions,
  tooltipPercentOptions,
} from "./strategies/per-range-tooltip-options";

/**
 * This function provides a compact, potentially lossy, humanized string representation of a number.
 * @param value The number to format
 * @param preset The format preset to use
 * @param type The format context type (e.g., tooltip, big-number)
 * @param locale Optional locale configuration for custom thousand/decimal separators
 */
export function humanizeDataType(
  value: number,
  preset: FormatPreset,
  type: FormatterContextSurface,
  locale?: LocaleConfig,
): string {
  if (typeof value !== "number") {
    console.warn(
      `humanizeDataType only accepts numbers, got ${value} for FormatPreset "${preset}"`,
    );
    return JSON.stringify(value);
  }

  const optionsMap: Record<FormatterContextSurface, ContextOptions> = {
    tooltip: {
      none: tooltipNoFormattingOptions,
      currencyUsd: tooltipCurrencyOptions(NumberKind.DOLLAR),
      currencyEur: tooltipCurrencyOptions(NumberKind.EURO),
      percent: tooltipPercentOptions,
      humanize: tooltipNoFormattingOptions,
    },
    "big-number": {
      none: bigNumDefaultFormattingOptions,
      currencyUsd: bigNumCurrencyOptions(NumberKind.DOLLAR),
      currencyEur: bigNumCurrencyOptions(NumberKind.EURO),
      percent: bigNumPercentOptions,
      humanize: bigNumDefaultFormattingOptions,
    },
    axis: {
      none: axisDefaultFormattingOptions,
      currencyUsd: axisCurrencyOptions(NumberKind.DOLLAR),
      currencyEur: axisCurrencyOptions(NumberKind.EURO),
      percent: axisPercentOptions,
      humanize: axisDefaultFormattingOptions,
    },
    table: {
      none: defaultNoFormattingOptions,
      currencyUsd: defaultCurrencyOptions(NumberKind.DOLLAR),
      currencyEur: defaultCurrencyOptions(NumberKind.EURO),
      percent: defaultPercentOptions,
      humanize: defaultGenericNumOptions,
    },
  };

  const selectedOptions = optionsMap[type] || optionsMap.table;

  switch (preset) {
    case FormatPreset.NONE:
      return new PerRangeFormatter(selectedOptions.none, locale).stringFormat(
        value,
      );

    case FormatPreset.CURRENCY_USD:
      return new PerRangeFormatter(
        selectedOptions.currencyUsd,
        locale,
      ).stringFormat(value);

    case FormatPreset.CURRENCY_EUR:
      return new PerRangeFormatter(
        selectedOptions.currencyEur,
        locale,
      ).stringFormat(value);

    case FormatPreset.PERCENTAGE:
      return new PerRangeFormatter(
        selectedOptions.percent,
        locale,
      ).stringFormat(value);

    case FormatPreset.INTERVAL:
      return type === "tooltip"
        ? formatMsToDuckDbIntervalString(value)
        : formatMsInterval(value);

    case FormatPreset.HUMANIZE:
      return new PerRangeFormatter(
        selectedOptions.humanize,
        locale,
      ).stringFormat(value);

    default:
      console.warn(
        "Unknown format preset, using none formatter. All number kinds should be handled.",
      );
      return new PerRangeFormatter(selectedOptions.none, locale).stringFormat(
        value,
      );
  }
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
    return formatMsToDuckDbIntervalString(value);
  }
  return value.toString();
}

const memoizedHumanizeDataType = memoize(humanizeDataType, {
  primitive: true,
  max: 1000,
});
const memoizedHumanizeDataTypeUnabridged = memoize(humanizeDataTypeUnabridged, {
  primitive: true,
});

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
  measureSpec: MetricsViewSpecMeasure,
  type: FormatterContext = "table",
): (value: number | string | T) => string | T {
  const useUnabridged = type === "unabridged";
  const isBigNumber = type === "big-number";
  const isAxis = type === "axis";
  const isTooltip = type === "tooltip";

  // Extract locale configuration from d3_locale
  const localeConfig: LocaleConfig | undefined =
    measureSpec.formatD3Locale && isValidD3Locale(measureSpec.formatD3Locale)
      ? {
          decimal: measureSpec.formatD3Locale.decimal as string | undefined,
          thousands: measureSpec.formatD3Locale.thousands as string | undefined,
          grouping: measureSpec.formatD3Locale.grouping as number[] | undefined,
        }
      : undefined;

  let humanizer: (value: number, type: FormatPreset) => string;
  if (useUnabridged) {
    humanizer = memoizedHumanizeDataTypeUnabridged;
  } else {
    humanizer = (value, preset) =>
      memoizedHumanizeDataType(value, preset, type, localeConfig);
  }

  // Return and empty string if measureSpec is not provided.
  // This may e.g. be the case during the initial render of a dashboard,
  // when a measureSpec has not yet loaded from a metadata query.
  if (measureSpec === undefined) {
    return (value: number | string | T) =>
      typeof value === "number" ? "" : value;
  }

  // Use the d3 formatter if it is provided and valid
  // (d3 throws an error if the format is invalid).
  // otherwise, use the humanize formatter.
  if (measureSpec.formatD3 !== undefined && measureSpec.formatD3 !== "") {
    try {
      let d3formatter: (n: number | { valueOf(): number }) => string;

      const isValidLocale = isValidD3Locale(measureSpec.formatD3Locale);
      if (isValidLocale) {
        const locale = getLocaleFromConfig(
          measureSpec.formatD3Locale as unknown as FormatLocaleDefinition,
        );
        d3formatter = d3FormatLocale(locale).format(measureSpec.formatD3);
      } else {
        d3formatter = d3format(measureSpec.formatD3);
      }

      const hasCurrencySymbol = includesCurrencySymbol(measureSpec.formatD3);
      const hasPercentSymbol = measureSpec.formatD3.includes("%");
      return (value: number | string | T) => {
        if (typeof value !== "number") return value;

        // For the Big Number, Axis and Tooltips, override the d3formatter
        // with humanized values that respect the locale configuration
        if (isBigNumber || isTooltip || isAxis) {
          if (hasCurrencySymbol) {
            if (isValidLocale && measureSpec?.formatD3Locale?.currency) {
              const currency = measureSpec.formatD3Locale.currency as [
                string,
                string,
              ];
              // Use custom currency symbol from locale
              const humanized = humanizer(value, FormatPreset.HUMANIZE);
              return `${currency[0]}${humanized}${currency[1]}`;
            }
            return humanizer(value, FormatPreset.CURRENCY_USD);
          } else if (hasPercentSymbol) {
            return humanizer(value, FormatPreset.PERCENTAGE);
          } else {
            return humanizer(value, FormatPreset.HUMANIZE);
          }
        }
        return d3formatter(value);
      };
    } catch (error) {
      console.warn("Invalid d3 format:", error);
      return (value: number | string | T) =>
        typeof value === "number"
          ? humanizer(value, FormatPreset.HUMANIZE)
          : value;
    }
  }

  // finally, use the formatPreset.
  let formatPreset =
    measureSpec.formatPreset && measureSpec.formatPreset !== ""
      ? (measureSpec.formatPreset as FormatPreset)
      : FormatPreset.NONE;

  if ((isAxis || isBigNumber) && formatPreset === FormatPreset.NONE) {
    formatPreset = FormatPreset.HUMANIZE;
  }

  return (value: number | T) =>
    typeof value === "number" ? humanizer(value, formatPreset) : value;
}
