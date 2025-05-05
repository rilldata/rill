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
import { removeSomeAdvancedMeasures } from "../state-managers/selectors/measures";

export function createTotalsForVisibleMeasures(
  ctx: StateManagers,
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const queryOptionsStore = derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.validSpecStore,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      ctx.selectors.measures.visibleMeasures,
    ],
    ([
      runtime,
      metricsViewName,
      validSpec,
      timeControls,
      dashboard,
      measures,
    ]) => {
      const { metricsView } = validSpec.data ?? {};

      const measuresToQuery = removeSomeAdvancedMeasures(
        dashboard,
        metricsView ?? {},
        measures.map((m) => m.name as string),
        false,
      );

      const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measuresToQuery.map((m) => ({
            name: m,
          })),
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
            enabled:
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              measuresToQuery.length > 0,
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

export function createUnfilteredTotalsForVisibleMeasures(
  ctx: StateManagers,
  dimensionName: string,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const queryOptionsStore = derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      ctx.selectors.measures.visibleMeasures,
    ],
    ([runtime, metricsViewName, timeControls, dashboard, measures]) => {
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
          measures: measures,
          where: updatedFilter,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
        },
        {
          query: {
            enabled:
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              measures.length > 0,
          },
        },
      );

      return queryOptions;
    },
  );

  const query = createQuery(queryOptionsStore);

  return query;
}
