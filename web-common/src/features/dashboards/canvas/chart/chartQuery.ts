import type { ChartTypeConfig } from "@rilldata/web-common/features/dashboards/canvas/types";
import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getChartData(ctx: StateManagers, config: ChartTypeConfig) {
  return derived(createChartDataQuery(ctx, config), (data) => {
    if (data?.data) {
      return data.data.data;
    } else {
      return [];
    }
  });
}

export function createChartDataQuery(
  ctx: StateManagers,
  config: ChartTypeConfig,
  limit = "500",
  offset = "0",
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  let measures: V1MetricsViewAggregationMeasure[] = [];
  if (config.data.y?.field) {
    measures = [{ name: config.data.y?.field }];
  }

  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.data.x?.field) {
    dimensions = [{ name: config.data.x?.field }];
  }
  if (config.data.color?.field) {
    dimensions = [...dimensions, { name: config.data.color.field }];
  }

  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          measures,
          dimensions,
          where: sanitiseExpression(mergeMeasureFilters(dashboard), undefined),
          timeRange: {
            start: timeControls.timeStart,
            end: timeControls.timeEnd,
          },
          limit,
          offset,
        },
        {
          query: {
            enabled: !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}
