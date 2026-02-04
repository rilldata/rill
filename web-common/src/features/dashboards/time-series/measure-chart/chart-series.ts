import type {
  TimeSeriesPoint,
  DimensionSeriesData,
  ChartSeries,
  ChartMode,
} from "./types";
import {
  MainLineColor,
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  TimeComparisonLineColor,
} from "../chart-colors";
import { COMPARISON_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";

const LINE_MODE_MIN_POINTS = 6;

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

export function determineMode(data: TimeSeriesPoint[]): ChartMode {
  return data.length >= LINE_MODE_MIN_POINTS ? "line" : "bar";
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
