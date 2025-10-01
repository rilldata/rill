import {
  CHART_CONFIG,
  type ChartSpec,
} from "@rilldata/web-common/features/canvas/components/charts";
import type {
  ChartDataResult,
  ChartDomainValues,
  ChartType,
  ColorMapping,
} from "@rilldata/web-common/features/components/charts/types";
import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";

export function generateSpec(
  chartType: ChartType,
  rillChartSpec: ChartSpec,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  return CHART_CONFIG[chartType]?.generateSpec(rillChartSpec, data);
}

export function isDomainStringArray(
  values: string[] | number[] | undefined,
): values is string[] {
  return values
    ? Array.isArray(values) &&
        values.every((value) => typeof value === "string")
    : false;
}

export function getColorForValues(
  colorValues: string[] | undefined,
  // if provided, use the colors for mentioned values
  overrideColorMapping: ColorMapping | undefined,
): ColorMapping | undefined {
  if (!colorValues || colorValues.length === 0) return undefined;

  const colorMapping = colorValues.map((value, index) => {
    const overrideColor = overrideColorMapping?.find(
      (mapping) => mapping.value === value,
    );
    return {
      value,
      color:
        overrideColor?.color ||
        COMPARIONS_COLORS[index % COMPARIONS_COLORS.length],
    };
  });

  return colorMapping;
}

export function getColorMappingForChart(
  chartSpec: ChartSpec,
  domainValues: ChartDomainValues | undefined,
): ColorMapping | undefined {
  if (!("color" in chartSpec) || !domainValues) return undefined;
  const colorField = chartSpec.color;

  let colorMapping: ColorMapping | undefined;
  if (typeof colorField === "object") {
    const fieldKey = colorField.field;
    const colorValues = domainValues[fieldKey];
    if (isDomainStringArray(colorValues)) {
      colorMapping = getColorForValues(colorValues, colorField.colorMapping);
    }
  }

  return colorMapping;
}
