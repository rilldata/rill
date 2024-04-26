import { measureFilterResolutionsStore } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
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
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function createTotalsForMeasure(
  ctx: StateManagers,
  measures: string[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      measureFilterResolutionsStore(ctx),
    ],
    (
      [
        runtime,
        metricsViewName,
        timeControls,
        dashboard,
        measureFilterResolution,
      ],
      set,
    ) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          where: sanitiseExpression(
            dashboard.whereFilter,
            measureFilterResolution.filter,
          ),
          timeStart: isComparison
            ? timeControls?.comparisonTimeStart
            : timeControls.timeStart,
          timeEnd: isComparison
            ? timeControls?.comparisonTimeEnd
            : timeControls.timeEnd,
        },
        {
          query: {
            enabled:
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              measureFilterResolution.ready,
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
      measureFilterResolutionsStore(ctx),
    ],
    (
      [
        runtime,
        metricsViewName,
        timeControls,
        dashboard,
        measureFilterResolution,
      ],
      set,
    ) => {
      const filter = sanitiseExpression(
        dashboard.whereFilter,
        measureFilterResolution.filter,
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
            enabled:
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              measureFilterResolution.ready,
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}
