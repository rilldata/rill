import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { activeMeasure } from "./active-measure";
import type { DashboardDataSources } from "./types";
import { visibleMeasures } from "./measures";

export const formattingSelectors = {
  /**
   * the currently active measure's format preset.
   */
  activeMeasureFormatPreset: (args: DashboardDataSources): FormatPreset =>
    (activeMeasure(args)?.formatPreset as FormatPreset) ??
    FormatPreset.HUMANIZE,

  /**
   * A readable containing a function that formats values
   * according to the active measure's format specification,
   * whether it's a d3 format string or a format preset.
   */
  activeMeasureFormatter: (args: DashboardDataSources) => {
    const measure = activeMeasure(args);
    if (measure === undefined) {
      return (_value: number | undefined) => undefined;
    }

    return createMeasureValueFormatter(measure);
  },

  /**
   * A map of measure names to their formatters
   */
  measureFormatters: (args: DashboardDataSources) => {
    const measures = visibleMeasures(args);
    return Object.fromEntries(
      measures.map((measure) => [
        measure.name,
        createMeasureValueFormatter(measure),
      ]),
    );
  },
};
