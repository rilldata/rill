import {
  FormatPreset,
  humanizeDataType,
  humanizeDataTypeExpanded,
} from "@rilldata/web-common/features/dashboards/humanize-numbers";
import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import { format as d3format } from "d3-format";

export const createMeasureValueFormatter = (
  measureSpec: MetricsViewSpecMeasureV2,
  useUnabridged = false
): ((value: number) => string) => {
  const humanizer = useUnabridged ? humanizeDataTypeExpanded : humanizeDataType;
  // if (value === undefined || value === null) return "";

  // Humanize by default if measureSpec is not provided.
  // This may e.g. be the case during the initial render of a dashboard.
  if (measureSpec === undefined)
    return (value: number) => humanizer(value, FormatPreset.HUMANIZE);

  // Use the d3 formatter if it is provided and valid
  // (d3 throws an error if the format is invalid).
  // otherwise, use the humanize formatter.
  if (measureSpec.formatD3 !== undefined && measureSpec.formatD3 !== "") {
    try {
      return d3format(measureSpec.formatD3);
    } catch (error) {
      return (value: number) => humanizer(value, FormatPreset.HUMANIZE);
    }
  }

  // finally, use the formatPreset.
  return (value: number) =>
    humanizer(
      value,
      (measureSpec.formatPreset as FormatPreset) ?? FormatPreset.HUMANIZE
    );
};
