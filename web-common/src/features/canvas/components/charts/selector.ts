import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import { useStartEndTime } from "@rilldata/web-common/features/canvas/components/kpi/selector";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getChartData(
  ctx: StateManagers,
  instanceId: string,
  config: ChartConfig,
) {
  return derived(createChartDataQuery(ctx, instanceId, config), (data) => {
    if (data?.data) {
      return data.data.data;
    } else {
      return [];
    }
  });
}

export function createChartDataQuery(
  ctx: StateManagers,
  instanceId: string,
  config: ChartConfig,
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

  return derived(
    [
      ctx.runtime,
      useStartEndTime(instanceId, config.metrics_view, config.time_range),
    ],
    ([runtime, timeRange], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        config.metrics_view,
        {
          measures,
          dimensions,
          where: undefined,
          timeRange: {
            start: timeRange?.data?.start?.toISOString() || undefined,
            end: timeRange?.data?.end?.toISOString() || undefined,
          },
          limit,
          offset,
        },
        {
          query: {
            enabled: !!timeRange.data,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}
