import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  filterExpressions,
  matchExpressionByName,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { V1MetricsViewAggregationResponse } from "@rilldata/web-common/runtime-client";
import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function createTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, Error> {
  return derived(
    [ctx.metricsViewName, useTimeControlStore(ctx), ctx.dashboardStore],
    ([metricsViewName, timeControls, dashboard], set) =>
      createQueryServiceMetricsViewAggregation(
        ctx.runtimeClient,
        {
          metricsView: metricsViewName,
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
            timeDimension: dashboard.selectedTimeDimension,
          },
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            refetchOnMount: false,
          },
        },
        ctx.queryClient,
      ).subscribe(set),
  );
}

export function createUnfilteredTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  dimensionName: string,
): CreateQueryResult<V1MetricsViewAggregationResponse, Error> {
  return derived(
    [ctx.metricsViewName, useTimeControlStore(ctx), ctx.dashboardStore],
    ([metricsViewName, timeControls, dashboard], set) => {
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

      createQueryServiceMetricsViewAggregation(
        ctx.runtimeClient,
        {
          metricsView: metricsViewName,
          measures: measures.map((measure) => ({ name: measure })),
          where: updatedFilter,
          timeRange: {
            start: timeControls.timeStart,
            end: timeControls.timeEnd,
            timeDimension: dashboard.selectedTimeDimension,
          },
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
          },
        },
        ctx.queryClient,
      ).subscribe(set);
    },
  );
}
