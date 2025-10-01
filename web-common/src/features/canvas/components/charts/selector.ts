import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/components/vega/util";
import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
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
import chroma from "chroma-js";
import { derived, type Readable } from "svelte/store";
import type { ChartDataResult, TimeDimensionDefinition } from "./types";
import { adjustDataForTimeZone, getFieldsByType } from "./util";

export function getChartData(
  ctx: CanvasStore,
  component: BaseChart<ChartSpec>,
  config: ChartSpec,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<ChartDataResult> {
  const chartDataQuery = component.createChartDataQuery(
    ctx,
    timeAndFilterStore,
  );

  const themeStore = ctx.canvasEntity.theme;

  const { measures, dimensions, timeDimensions } = getFieldsByType(config);

  // Combine all fields with their types
  const allFields = [
    ...measures.map((field) => ({ field, type: "measure" })),
    ...dimensions.map((field) => ({ field, type: "dimension" })),
    ...timeDimensions.map((field) => ({ field, type: "time" })),
  ];

  // Match each field to its corresponding measure or dimension spec.
  const metricsView = ctx.canvasEntity.metricsView;

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

      const domainValues = component.getChartDomainValues();

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
