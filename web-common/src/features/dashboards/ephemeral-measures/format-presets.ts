import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

/**
 * Format presets offered when defining an ephemeral measure.
 */
export function ephemeralFormatPresetOptions(): {
  value: string;
  label: string;
}[] {
  return [
    {
      value: FormatPreset.HUMANIZE,
      label: m.dashboard_pivot_ephemeral_format_humanize(),
    },
    { value: FormatPreset.NONE, label: m.common_none() },
    {
      value: FormatPreset.CURRENCY_USD,
      label: m.dashboard_pivot_ephemeral_format_currency_usd(),
    },
    {
      value: FormatPreset.CURRENCY_EUR,
      label: m.dashboard_pivot_ephemeral_format_currency_eur(),
    },
    {
      value: FormatPreset.PERCENTAGE,
      label: m.dashboard_pivot_ephemeral_format_percentage(),
    },
  ];
}
