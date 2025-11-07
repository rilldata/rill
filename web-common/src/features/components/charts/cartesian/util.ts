/**
 * Determines the pivot configuration for the bar chart based on the presence of
 * x/y fields, color fields, and comparison mode.
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
      field: hasComparison ? "color_with_comparison" : colorField,
      value: yField,
      groupby: [xField],
    };
  }

  // In comparison mode without color field, pivot on measure_key
  if (hasComparison) {
    return {
      field: "measure_key",
      value: yField,
      groupby: [xField],
    };
  }

  return undefined;
}
