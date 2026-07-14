import { COMPARISON_COLORS } from "@rilldata/web-common/features/dashboards/config";
import {
  isAdaptiveChartType,
  TDDChart,
} from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
  TimeComparisonLineColor,
} from "../chart-colors";
import { LINE_MODE_MIN_POINTS } from "./scales";
import type {
  ChartMode,
  ChartSeries,
  DimensionSeriesData,
  TimeSeriesPoint,
} from "./types";

export function buildChartSeries(
  data: TimeSeriesPoint[],
  dimData: DimensionSeriesData[],
  showComparison: boolean,
): ChartSeries[] {
  if (dimData.length > 0) {
    return dimData.map((dim, i) => ({
      id: `dim-${dim.dimensionValue ?? i}`,
      values: dim.data.map((pt) => pt.value),
      color: dim.color || COMPARISON_COLORS[i % COMPARISON_COLORS.length],
      opacity: dim.isFetching ? 0.5 : 1,
      strokeWidth: 1.5,
    }));
  }

  const result: ChartSeries[] = [];

  if (data.length > 0) {
    result.push({
      id: "primary",
      values: data.map((pt) => pt.value),
      color: MainLineColor,
      areaGradient: {
        dark: MainAreaColorGradientDark,
        light: MainAreaColorGradientLight,
      },
    });
  }

  if (showComparison && data.length > 0) {
    result.push({
      id: "comparison",
      values: data.map((pt) => pt.comparisonValue ?? null),
      color: TimeComparisonLineColor,
      opacity: 0.5,
    });
  }

  return result;
}

export function determineMode(
  chartType: TDDChart,
  data: TimeSeriesPoint[],
): ChartMode {
  if (chartType === TDDChart.LINE) {
    return "line";
  }
  return data.length >= LINE_MODE_MIN_POINTS ? "line" : "bar";
}

export function resolveEffectiveChartType(
  chartType: TDDChart,
  hasDimensionComparison: boolean,
): TDDChart {
  if (isAdaptiveChartType(chartType) && hasDimensionComparison) {
    return TDDChart.STACKED_BAR;
  }
  return chartType;
}

export function usesVegaRenderer(effectiveChartType: TDDChart): boolean {
  return (
    !isAdaptiveChartType(effectiveChartType) &&
    effectiveChartType !== TDDChart.LINE
  );
}

export function computeTooltipDelta(point: TimeSeriesPoint | null) {
  const currentValue = point?.value ?? null;
  const comparisonValue = point?.comparisonValue ?? null;
  const delta =
    currentValue !== null && comparisonValue !== null && comparisonValue !== 0
      ? (currentValue - comparisonValue) / comparisonValue
      : null;
  return {
    currentValue,
    comparisonValue,
    deltaLabel:
      delta !== null
        ? numberPartsToString(formatMeasurePercentageDifference(delta))
        : null,
    deltaPositive: delta !== null && delta >= 0,
  };
}
