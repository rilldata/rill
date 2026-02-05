import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { type ChartType } from "../../../components/charts/types";
import { type CanvasChartSpec } from "./";

const allowedTimeDimensionDetailTypes = [
  "line_chart",
  "area_chart",
  "stacked_bar",
  "stacked_bar_normalized",
  "bar_chart",
];

export const CanvasChartTypeToTDDChartType = {
  line_chart: TDDChart.DEFAULT,
  area_chart: TDDChart.STACKED_AREA,
  stacked_bar: TDDChart.STACKED_BAR,
  stacked_bar_normalized: TDDChart.STACKED_BAR,
  bar_chart: TDDChart.GROUPED_BAR,
};

export function getLinkStateForTimeDimensionDetail(
  spec: CanvasChartSpec,
  type: ChartType,
): {
  canLink: boolean;
  measureName?: string;
  dimensionName?: string;
} {
  if (!allowedTimeDimensionDetailTypes.includes(type))
    return {
      canLink: false,
    };

  const hasXAxis = "x" in spec;
  const hasYAxis = "y" in spec;
  if (!hasXAxis || !hasYAxis)
    return {
      canLink: false,
    };

  const xAxis = spec.x;
  const yAxis = spec.y;

  if (isFieldConfig(xAxis) && isFieldConfig(yAxis)) {
    const colorDimension = spec.color;
    if (isFieldConfig(colorDimension)) {
      return {
        canLink: xAxis.type === "temporal",
        measureName: yAxis.field,
        dimensionName: colorDimension.field,
      };
    }

    return {
      canLink: xAxis.type === "temporal",
      measureName: yAxis.field,
    };
  }
  return {
    canLink: false,
  };
}
