import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  type RpcStatus,
  type V1MetricsViewComparisonResponse,
  type V1MetricsViewSpec,
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

export const useMetricsView = <T = V1MetricsViewSpec>(
  ctx: StateManagers,
  selector?: (meta: V1MetricsViewSpec) => T,
): Readable<QueryObserverResult<T | V1MetricsViewSpec, RpcStatus>> => {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricViewName], set) => {
      return useResource(
        runtime.instanceId,
        metricViewName,
        ResourceKind.MetricsView,
        {
          select: (data) =>
            selector
              ? selector(data.resource?.metricsView?.state?.validSpec)
              : data.resource?.metricsView?.state?.validSpec,
          queryClient: ctx.queryClient,
        },
      ).subscribe(set);
    },
  );
};

export const useModelHasTimeSeries = (ctx: StateManagers) =>
  useMetricsView(
    ctx,
    (meta) => !!meta?.timeDimension,
  ) as CreateQueryResult<boolean>;

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
    ([metricsExplorer, timeControls, metricViewName, runtime], set) => {
      return createQueryServiceMetricsViewComparison(
        runtime.instanceId,
        metricViewName,
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
    [ctx.runtime, ctx.metricsViewName, useMetricsView(ctx)],
    ([runtime, metricsViewName, metricsView], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        metricsViewName,
        {},
        {
          query: {
            queryClient: ctx.queryClient,
            enabled: !metricsView.error && !!metricsView.data?.timeDimension,
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
