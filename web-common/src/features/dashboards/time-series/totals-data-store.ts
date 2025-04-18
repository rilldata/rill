import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  filterExpressions,
  matchExpressionByName,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import { type HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function createTotalsForMeasures(
  ctx: StateManagers,
  measures: string[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const queryOptionsStore = derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard]) => {
      const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          where: sanitiseExpression(
            mergeDimensionAndMeasureFilters(
              dashboard.whereFilter,
              dashboard.dimensionThresholdFilters,
            ),
            undefined,
          ),
          timeRange: {
            start: isComparison
              ? timeControls?.comparisonTimeStart
              : timeControls.timeStart,
            end: isComparison
              ? timeControls?.comparisonTimeEnd
              : timeControls.timeEnd,
          },
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            refetchOnMount: false,
          },
        },
      );

      return queryOptions;
    },
  );

  const query = createQuery(queryOptionsStore);

  return query;
}

export function createUnfilteredTotalsForMeasures(
  ctx: StateManagers,
  measures: string[],
  dimensionName: string,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const queryOptionsStore = derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard]) => {
      const filter = sanitiseExpression(
        mergeDimensionAndMeasureFilters(
          dashboard.whereFilter,
          dashboard.dimensionThresholdFilters,
        ),
        undefined,
      );

      const updatedFilter = filterExpressions(
        filter || createAndExpression([]),
        (e) => !matchExpressionByName(e, dimensionName),
      );

      const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          where: updatedFilter,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
          },
        },
      );

      return queryOptions;
    },
  );

  const query = createQuery(queryOptionsStore);

  return query;
}
