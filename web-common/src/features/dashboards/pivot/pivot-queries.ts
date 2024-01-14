import {
  createQueryServiceMetricsViewAggregation,
  V1MetricsViewFilter,
  type V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
  V1BuiltinMeasure,
} from "@rilldata/web-common/runtime-client";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, writable } from "svelte/store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";

/**
 * Wrapper function for Aggregate Query API
 */
export function createPivotAggregationRowQuery(
  ctx: StateManagers,
  measures: string[],
  dimensions: string[],
  filters: V1MetricsViewFilter,
  sort: V1MetricsViewAggregationSort[] = [],
  limit = "100",
  offset = "0",
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  // Todo: Handle sorting in table
  if (!sort.length) {
    sort = [
      {
        desc: false,
        name: measures[0] || dimensions[0],
      },
    ];
  }
  return derived(
    [ctx.runtime, ctx.metricsViewName, useTimeControlStore(ctx)],
    ([runtime, metricViewName, timeControls], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          dimensions: dimensions.map((dimension) => ({ name: dimension })),
          filter: filters,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          sort,
          limit,
          offset,
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}

/***
 * Get a list of axis values for a given list of dimension values and filters
 */
export function getAxisForDimensions(
  ctx: StateManagers,
  dimensions: string[],
  filters: V1MetricsViewFilter,
  sortBy: V1MetricsViewAggregationSort[] = [],
) {
  if (!dimensions.length) return writable(null);

  // FIXME: If sorting by measure, add that to measure list
  let measures: string[] = [];
  if (sortBy.length) {
    const sortMeasure = sortBy[0].name as string;
    if (!dimensions.includes(sortMeasure)) {
      measures = [sortMeasure];
    }
  }

  return derived(
    dimensions.map((dimension) =>
      createPivotAggregationRowQuery(
        ctx,
        measures,
        [dimension],
        filters,
        sortBy,
      ),
    ),
    (data) => {
      const axesMap: Record<string, string[]> = {};

      // Wait for all data to populate
      if (data.some((d) => d?.isFetching)) return { isFetching: true };

      data.forEach((d, i: number) => {
        const dimensionName = dimensions[i];
        axesMap[dimensionName] = (d?.data?.data || [])?.map(
          (dimValue) => dimValue[dimensionName] as string,
        );
      });

      if (Object.values(axesMap).some((d) => !d)) return { isFetching: true };

      return {
        isFetching: false,
        data: axesMap,
      };
    },
  );
}

/**
 * Get a count of unique values for a given dimension
 */
export function getDimensionCount(
  ctx: StateManagers,
  dimensionName: string,
  filters: V1MetricsViewFilter,
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  const measures = [
    {
      name: "__count",
      builtinMeasure: V1BuiltinMeasure.BUILTIN_MEASURE_COUNT_DISTINCT,
      builtinMeasureArgs: [dimensionName],
    },
  ];
  return derived(
    [ctx.runtime, ctx.metricsViewName, useTimeControlStore(ctx)],
    ([runtime, metricViewName, timeControls], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricViewName,
        {
          measures,
          filter: filters,
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}
