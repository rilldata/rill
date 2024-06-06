import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  filterExpressions,
  matchExpressionByName,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createQueryServiceMetricsViewAggregation,
  MetricsViewSpecDimensionSelector,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function createTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  dimensions: MetricsViewSpecDimensionSelector[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse> {
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
          measures: measures.map((measure) => ({ name: measure })),
          dimensions,
          where: sanitiseExpression(mergeMeasureFilters(dashboard), undefined),
          timeStart: isComparison
            ? timeControls?.comparisonTimeStart
            : timeControls.timeStart,
          timeEnd: isComparison
            ? timeControls?.comparisonTimeEnd
            : timeControls.timeEnd,
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set),
  );
}

export function createUnfilteredTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  dimensionName: string,
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard], set) => {
      const filter = sanitiseExpression(
        mergeMeasureFilters(dashboard),
        undefined,
      );

      const updatedFilter = filterExpressions(
        filter || createAndExpression([]),
        (e) => !matchExpressionByName(e, dimensionName),
      );

      createQueryServiceMetricsViewAggregation(
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
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}
