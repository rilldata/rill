import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/components/vega/util";
import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { MetricsViewSelectors } from "@rilldata/web-common/features/canvas/stores/metrics-view-selectors";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import {
  defaultPrimaryColors,
  defaultSecondaryColors,
} from "@rilldata/web-common/features/themes/color-config";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import { derived, type Readable } from "svelte/store";
import type {
  ChartDataQuery,
  ChartDataResult,
  ChartDomainValues,
  TimeDimensionDefinition,
} from "../../../components/charts/types";
import { adjustDataForTimeZone, getFieldsByType } from "./util";

export interface ChartDataDependencies<T extends ChartSpec = ChartSpec> {
  config: T;
  chartDataQuery: ChartDataQuery;
  metricsView: MetricsViewSelectors;
  themeStore: Readable<{ primary?: Color; secondary?: Color }>;
  timeAndFilterStore: Readable<TimeAndFilterStore>;
  getDomainValues: () => ChartDomainValues;
}

/**
 * Pure function to create a chart data store without canvas context dependency.
 * This can be used outside of canvas context by providing all dependencies explicitly.
 */
export function getChartData<T extends ChartSpec = ChartSpec>(
  deps: ChartDataDependencies<T>,
): Readable<ChartDataResult> {
  const {
    config,
    chartDataQuery,
    metricsView,
    themeStore,
    getDomainValues,
    timeAndFilterStore,
  } = deps;

  const { measures, dimensions, timeDimensions } = getFieldsByType(config);

  // Combine all fields with their types
  const allFields = [
    ...measures.map((field) => ({ field, type: "measure" })),
    ...dimensions.map((field) => ({ field, type: "dimension" })),
    ...timeDimensions.map((field) => ({ field, type: "time" })),
  ];

  const fieldReadableMap = allFields.map((field) => {
    if (field.type === "measure") {
      return metricsView.getMeasureForMetricView(
        field.field,
        config.metrics_view,
      );
    } else if (field.type === "dimension") {
      return metricsView.getDimensionForMetricView(
        field.field,
        config.metrics_view,
      );
    } else {
      return getTimeDimensionDefinition(field.field, timeAndFilterStore);
    }
  });

  return derived(
    [chartDataQuery, timeAndFilterStore, themeStore, ...fieldReadableMap],
    ([chartData, $timeAndFilterStore, theme, ...fieldMap]) => {
      const fieldSpecMap = allFields.reduce(
        (acc, field, index) => {
          acc[field.field] = fieldMap?.[index];
          return acc;
        },
        {} as Record<
          string,
          | MetricsViewSpecMeasure
          | MetricsViewSpecDimension
          | TimeDimensionDefinition
          | undefined
        >,
      );

      let data = chartData?.data?.data;

      if (timeDimensions?.length && $timeAndFilterStore.timeGrain) {
        data = adjustDataForTimeZone(
          data,
          timeDimensions,
          $timeAndFilterStore.timeGrain,
          $timeAndFilterStore.timeRange.timeZone || "UTC",
        );
      }

      const domainValues = getDomainValues();

      return {
        data: data || [],
        isFetching: chartData?.isFetching ?? false,
        error: chartData?.error,
        fields: fieldSpecMap,
        domainValues,
        theme: {
          primary: theme.primary || chroma(`hsl(${defaultPrimaryColors[500]})`),
          secondary:
            theme.secondary || chroma(`hsl(${defaultSecondaryColors[500]})`),
        },
      };
    },
  );
}

/**
 * Convenience wrapper for using getChartData with canvas context.
 */
export function getChartDataForCanvas(
  ctx: CanvasStore,
  component: BaseChart<ChartSpec>,
  config: ChartSpec,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<ChartDataResult> {
  const chartDataQuery = component.createChartDataQuery(
    ctx,
    timeAndFilterStore,
  );

  return getChartData({
    config,
    chartDataQuery,
    getDomainValues: () => component.getChartDomainValues(),
    metricsView: ctx.canvasEntity.metricsView,
    themeStore: ctx.canvasEntity.theme,
    timeAndFilterStore,
  });
}

export function getTimeDimensionDefinition(
  field: string,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<TimeDimensionDefinition> {
  return derived(timeAndFilterStore, ($timeAndFilterStore) => {
    const grain = $timeAndFilterStore?.timeGrain;
    const displayName = "Time";

    if (grain) {
      const timeUnit = timeGrainToVegaTimeUnitMap[grain];
      const format = TIME_GRAIN[grain]?.d3format as string;
      return {
        field,
        timeUnit,
        displayName,
        format,
      };
    }
    return {
      field,
      displayName,
    };
  });
}
