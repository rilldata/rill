import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import type {
  ChartSpec,
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
): CartesianChartSpec & Pick<ChartSpec, "vl_config"> {
  const spec: CartesianChartSpec & Pick<ChartSpec, "vl_config"> = {
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
    // Fix vertical alignment across stacked TDD charts: force a fixed axis
    // width so every measure's plot area is identical regardless of label width.
    vl_config: JSON.stringify({
      axisRight: {
        minExtent: 50,
        maxExtent: 50,
      },
    }),
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
