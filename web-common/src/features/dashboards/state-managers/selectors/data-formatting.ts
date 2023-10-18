import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "../../humanize-numbers";
import type { SelectorFnArgs } from "./types";
import { activeMeasure } from "./active-measure";

export const formattingSelectors = {
  /**
   * the currently active measure's format preset.
   */
  activeMeasureFormatPreset: (args: SelectorFnArgs): FormatPreset =>
    (activeMeasure(args)?.formatPreset as FormatPreset) ??
    FormatPreset.HUMANIZE,

  /**
   * A readable containing a function that formats values
   * according to the active measure's format specification,
   * whether it's a d3 format string or a format preset.
   */
  activeMeasureFormatter: (args: SelectorFnArgs) =>
    createMeasureValueFormatter(activeMeasure(args)),
};
