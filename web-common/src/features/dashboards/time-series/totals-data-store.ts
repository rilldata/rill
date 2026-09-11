import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { V1MetricsViewAggregationResponse } from "@rilldata/web-common/runtime-client";
import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import { getFiltersForOtherDimensions } from "@rilldata/web-common/features/dashboards/selectors.ts";

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
          where: sanitiseExpression(dashboard.whereFilter, undefined),
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
      const filter = sanitiseExpression(dashboard.whereFilter, undefined);

      const updatedFilter = getFiltersForOtherDimensions(filter, dimensionName);

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
