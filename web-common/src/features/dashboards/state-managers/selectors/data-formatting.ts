import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import type { DashboardDataSources } from "./types";
import { activeMeasure } from "./active-measure";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

export const formattingSelectors = {
  /**
   * the currently active measure's format preset. If no measure is active,
   * or the active measure has no format preset, this will return the
   * default format preset, FormatPreset.HUMANIZE.
   */
  activeMeasureFormatPreset: (args: DashboardDataSources): FormatPreset =>
    (activeMeasure(args)?.formatPreset as FormatPreset) ??
    FormatPreset.HUMANIZE,

  /**
   * A readable containing a function that formats values
   * according to the active measure's format specification,
   * whether it's a d3 format string or a format preset.
   *
   * Note that this formatter is ONLY valid when an active measure
   * is present. If no measure is active, the formatter
   * will return an empty string.
   */
  activeMeasureFormatter: (args: DashboardDataSources) => {
    const measure = activeMeasure(args);
    if (measure) {
      createMeasureValueFormatter(measure);
    }

    // allowed to make make function signatures match in both cases
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    return (value: number) => "";
  },
};
