import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import {
  ColorWithComparisonField,
  MeasureKeyField,
} from "@rilldata/web-common/features/components/charts/comparison-builder";

/**
 * Determines the pivot configuration for the bar chart based on the presence of
 * x/y fields, color fields, and comparison mode.
 *
 * `xField` and `yField` are already escaped Vega field references. `colorField`
 * is the raw dimension name and is escaped here: Vega resolves the pivot field
 * as a field path, so a name with dots such as `query.context.metaData.type`
 * would otherwise be read as a nested lookup and throw inside the dataflow,
 * leaving the whole chart blank.
 */
export function createVegaTransformPivotConfig(
  xField: string | undefined,
  yField: string | undefined,
  colorField: string | undefined,
  hasComparison: boolean,
  hasMultiValueTooltip: boolean,
) {
  // No pivot if we don't have x, y fields and multi-value tooltips
  if (!xField || !yField || !hasMultiValueTooltip) {
    return undefined;
  }

  if (colorField) {
    return {
      // Use color_with_comparison field when in comparison mode to include both current and previous values
      field: hasComparison
        ? ColorWithComparisonField
        : sanitizeValueForVega(colorField),
      value: yField,
      groupby: [xField],
    };
  }

  // In comparison mode without color field, pivot on measure_key
  if (hasComparison) {
    return {
      field: MeasureKeyField,
      value: yField,
      groupby: [xField],
    };
  }

  return undefined;
}
