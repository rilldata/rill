import {
  applyExactMeasures,
  mapEphemeralMeasuresForRequest,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import { requestExactNames } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import { exploreExactSurface } from "@rilldata/web-common/features/dashboards/exact-measures/explore-surfaces";
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
  // The surface whose exact selection the request follows; see `exploreExactSurface`.
  exactSurface: string = exploreExactSurface.measures,
): CreateQueryResult<V1MetricsViewAggregationResponse, Error> {
  return derived(
    [
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      ctx.selectors.exactMeasures.exactMeasureContext,
    ],
    ([metricsViewName, timeControls, dashboard, exactContext], set) =>
      createQueryServiceMetricsViewAggregation(
        ctx.runtimeClient,
        {
          metricsView: metricsViewName,
          measures: applyExactMeasures(
            mapEphemeralMeasuresForRequest(
              measures.map((measure) => ({ name: measure })),
              dashboard.ephemeralMeasures,
            ),
            requestExactNames(measures, exactContext(exactSurface)),
          ),
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
  // The surface whose exact selection the request follows; see `exploreExactSurface`.
  exactSurface: string = exploreExactSurface.measures,
): CreateQueryResult<V1MetricsViewAggregationResponse, Error> {
  return derived(
    [
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      ctx.selectors.exactMeasures.exactMeasureContext,
    ],
    ([metricsViewName, timeControls, dashboard, exactContext], set) => {
      const filter = sanitiseExpression(dashboard.whereFilter, undefined);

      const updatedFilter = getFiltersForOtherDimensions(filter, dimensionName);

      return createQueryServiceMetricsViewAggregation(
        ctx.runtimeClient,
        {
          metricsView: metricsViewName,
          measures: applyExactMeasures(
            mapEphemeralMeasuresForRequest(
              measures.map((measure) => ({ name: measure })),
              dashboard.ephemeralMeasures,
            ),
            requestExactNames(measures, exactContext(exactSurface)),
          ),
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
