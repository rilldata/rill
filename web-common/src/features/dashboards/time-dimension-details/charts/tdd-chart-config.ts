import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import type {
  ChartType,
  ColorMapping,
} from "@rilldata/web-common/features/components/charts/types";
import type { DimensionSeriesData } from "@rilldata/web-common/features/dashboards/time-series/measure-chart/types";
import { TDDChart } from "../types";

/**
 * Maps TDD chart types to Canvas chart types.
 */
export const TDD_TO_COMPONENT_CHART_TYPE: Record<TDDChart, ChartType> = {
  [TDDChart.DEFAULT]: "line_chart",
  [TDDChart.STACKED_BAR]: "stacked_bar",
  [TDDChart.GROUPED_BAR]: "bar_chart",
  [TDDChart.STACKED_AREA]: "area_chart",
};

function buildColorMapping(dimData: DimensionSeriesData[]): ColorMapping {
  return dimData
    .filter((d) => d.dimensionValue !== null)
    .map((d) => ({
      value: d.dimensionValue as string,
      color: d.color,
    }));
}

/**
 * Creates a CartesianChartSpec from TDD parameters for use with component charts.
 */
export function createTDDCartesianSpec(
  metricsViewName: string,
  measureName: string,
  timeDimension: string,
  comparisonDimension?: string,
  selectedValues?: (string | null)[],
  dimensionData?: DimensionSeriesData[],
  showTimeDimensionDetail = true,
): CartesianChartSpec {
  const spec: CartesianChartSpec = {
    metrics_view: metricsViewName,
    x: {
      field: timeDimension,
      type: "temporal",
      axisOrient: showTimeDimensionDetail ? "top" : "none",
    },
    y: {
      field: measureName,
      type: "quantitative",
      axisOrient: "right",
    },
    isInteractive: true,
  };

  const colorMapping = dimensionData
    ? buildColorMapping(dimensionData)
    : undefined;

  // Add color dimension for dimension comparison
  if (comparisonDimension && selectedValues?.length) {
    const filteredValues = selectedValues.filter(
      (v): v is string => v !== null,
    );
    spec.color = {
      field: comparisonDimension,
      type: "nominal",
      legendOrientation: "none",
      values: filteredValues,
      ...(colorMapping?.length ? { colorMapping } : {}),
    };
  } else {
    spec.color = "primary";
  }

  return spec;
}
