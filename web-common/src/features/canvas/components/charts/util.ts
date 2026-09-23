import { toVerticalSpec } from "@rilldata/web-common/features/components/charts/cartesian/orientation";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { type ChartType } from "../../../components/charts/types";
import { type CanvasChartSpec } from "./";
import type { CartesianCanvasChartSpec } from "./variants/CartesianChart";

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
    return { canLink: false };

  const hasXAxis = "x" in spec;
  const hasYAxis = "y" in spec;
  if (!hasXAxis || !hasYAxis) return { canLink: false };

  // Only cartesian types pass the gate above; a horizontal bar chart carries
  // its measure on x, so read the fields in the vertical layout.
  const { x: xAxis, y: yAxis } = toVerticalSpec(
    spec as CartesianCanvasChartSpec,
  );

  if (!isFieldConfig(xAxis) || !isFieldConfig(yAxis)) return { canLink: false };

  if (yAxis.fields && yAxis.fields.length > 1) return { canLink: false };

  const colorDimension = spec.color;
  const hasDimensionBreakout =
    isFieldConfig(colorDimension) &&
    colorDimension.type !== "quantitative" &&
    colorDimension.type !== "value";

  if (hasDimensionBreakout && xAxis.type === "nominal")
    return { canLink: false };

  if (hasDimensionBreakout) {
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
