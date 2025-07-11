import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
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
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import { type HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  createQuery,
  keepPreviousData,
  type CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function createTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  includeComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const queryOptionsStore = derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard]) => {
      const measuresList: V1MetricsViewAggregationMeasure[] = measures.flatMap(
        (measureName) => {
          const baseMeasure = { name: measureName };
          if (!includeComparison) {
            return [baseMeasure];
          }
          return [
            baseMeasure,
            {
              name: measureName + ComparisonDeltaPreviousSuffix,
              comparisonValue: { measure: measureName },
            },
          ];
        },
      );

      const enabled =
        !!timeControls.ready &&
        !!ctx.dashboardStore &&
        // in case of comparison, we need to wait for the comparison start time to be available
        (!includeComparison || !!timeControls.comparisonTimeStart);

      return getQueryServiceMetricsViewAggregationQueryOptions(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measuresList,
          where: sanitiseExpression(
            mergeDimensionAndMeasureFilters(
              dashboard.whereFilter,
              dashboard.dimensionThresholdFilters,
            ),
            undefined,
          ),
          timeRange: {
            start: timeControls.timeStart,
            end: timeControls.timeEnd,
          },
          comparisonTimeRange: includeComparison
            ? {
                start: timeControls?.comparisonTimeStart,
                end: timeControls?.comparisonTimeEnd,
              }
            : undefined,
        },
        {
          query: {
            enabled,
            refetchOnMount: false,
            placeholderData: keepPreviousData,
          },
        },
      );
    },
  );

  return createQuery(queryOptionsStore, ctx.queryClient);
}

export function createUnfilteredTotalsForMeasure(
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

      const enabled = !!timeControls.ready && !!ctx.dashboardStore;

      return getQueryServiceMetricsViewAggregationQueryOptions(
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
            enabled,
            placeholderData: keepPreviousData,
          },
        },
      );
    },
  );

  return createQuery(queryOptionsStore, ctx.queryClient);
}
