import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  createQueryServiceMetricsViewAggregation,
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type ChartDataResult = {
  data: V1MetricsViewAggregationResponseDataItem[];
  isFetching: boolean;
  fields: Record<
    string,
    MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2 | undefined
  >;
  error?: HTTPError | null;
};

export function getChartData(
  ctx: StateManagers,
  instanceId: string,
  config: ChartConfig,
): Readable<ChartDataResult> {
  const chartDataQuery = createChartDataQuery(ctx, config);
  const { spec } = ctx.canvasEntity;

  const fields: { name: string; type: "measure" | "dimension" }[] = [];
  if (config.y?.field) fields.push({ name: config.y.field, type: "measure" });
  if (config.x?.field) fields.push({ name: config.x.field, type: "dimension" });
  if (typeof config.color === "object" && config.color?.field) {
    fields.push({ name: config.color.field, type: "dimension" });
  }

  // Match each field to its corresponding measure or dimension spec.
  const fieldReadableMap = fields.map((field) => {
    if (field.type === "measure") {
      return spec.getMeasureForMetricView(field.name, config.metrics_view);
    } else {
      return spec.getDimensionForMetricView(field.name, config.metrics_view);
    }
  });

  return derived(
    [chartDataQuery, ...fieldReadableMap],
    ([chartData, ...fieldMap]) => {
      const fieldSpecMap = fields.reduce(
        (acc, field, index) => {
          acc[field.name] = fieldMap?.[index];
          return acc;
        },
        {} as Record<
          string,
          MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2 | undefined
        >,
      );
      return {
        data: chartData?.data?.data || [],
        isFetching: chartData.isFetching,
        error: chartData.error,
        fields: fieldSpecMap,
      };
    },
  );
}

export function createChartDataQuery(
  ctx: StateManagers,
  config: ChartConfig & ComponentFilterProperties,
  limit = "500",
  offset = "0",
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  let measures: V1MetricsViewAggregationMeasure[] = [];
  if (config.y?.field) {
    measures = [{ name: config.y?.field }];
  }

  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.x?.field) {
    dimensions = [{ name: config.x?.field }];
  }
  if (typeof config.color === "object" && config.color?.field) {
    dimensions = [...dimensions, { name: config.color.field }];
  }

  const { timeControls } = ctx.canvasEntity;
  return derived(
    [ctx.runtime, timeControls.selectedTimeRange],
    ([runtime, selectedTimeRange], set) => {
      let timeRange: V1TimeRange = {
        start: selectedTimeRange?.start?.toISOString(),
        end: selectedTimeRange?.end?.toISOString(),
      };

      if (config.time_range) {
        timeRange = { isoDuration: config.time_range };
      }
      return createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        config.metrics_view,
        {
          measures,
          dimensions,
          where: undefined,
          timeRange,
          limit,
          offset,
        },
        {
          query: {
            enabled: !!selectedTimeRange?.start && !!selectedTimeRange?.end,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set);
    },
  );
}
