import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  type RpcStatus,
  type V1MetricsViewComparisonResponse,
  type V1MetricsViewTimeRangeResponse,
  createQueryServiceMetricsViewComparison,
  createQueryServiceMetricsViewSchema,
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewSchemaResponse,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";
import { type Readable, derived } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";

export const getFilterSearchList = (
  ctx: StateManagers,
  {
    dimension,
    addNull,
    searchText,
  }: {
    dimension: string;
    addNull: boolean;
    searchText: string;
  },
): Readable<
  QueryObserverResult<V1MetricsViewComparisonResponse, RpcStatus>
> => {
  return derived(
    [
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      ctx.metricsViewName,
      ctx.runtime,
    ],
    ([metricsExplorer, timeControls, metricsViewName, runtime], set) => {
      return createQueryServiceMetricsViewComparison(
        runtime.instanceId,
        metricsViewName,
        {
          dimension: { name: dimension },
          measures: [{ name: metricsExplorer.leaderboardMeasureName }],
          timeRange: {
            start: timeControls.timeStart,
            end: timeControls.timeEnd,
          },
          limit: "100",
          offset: "0",
          sort: [{ name: dimension }],
          where: addNull
            ? createInExpression(dimension, [null])
            : createLikeExpression(dimension, `%${searchText}%`),
        },
        {
          query: {
            queryClient: ctx.queryClient,
            enabled: timeControls.ready,
          },
        },
      ).subscribe(set);
    },
  );
};

export function createTimeRangeSummary(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewTimeRangeResponse> {
  return derived(
    [ctx.runtime, ctx.metricsViewName, ctx.validSpecStore],
    ([runtime, metricsViewName, validSpec], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        metricsViewName,
        {},
        {
          query: {
            queryClient: ctx.queryClient,
            enabled:
              !validSpec.error && !!validSpec.data?.metricsView?.timeDimension,
          },
        },
      ).subscribe(set),
  );
}

export function createMetricsViewSchema(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewSchemaResponse> {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricsViewName], set) =>
      createQueryServiceMetricsViewSchema(
        runtime.instanceId,
        metricsViewName,
        {},
        {
          query: {
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set),
  );
}
