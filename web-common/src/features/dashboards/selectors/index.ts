import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { STRING_LIKES } from "@rilldata/web-common/lib/duckdb-data-types";
import {
  RpcStatus,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
  createQueryServiceMetricsViewTimeRange,
  createQueryServiceMetricsViewSchema,
  type V1MetricsViewSchemaResponse,
  createQueryServiceMetricsViewComparison,
  V1MetricsViewComparisonResponse,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";
import { Readable, derived } from "svelte/store";
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
        (data) =>
          selector
            ? selector(data.metricsView?.state?.validSpec)
            : data.metricsView?.state?.validSpec,
        ctx.queryClient,
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
    type,
  }: {
    dimension: string;
    addNull: boolean;
    searchText: string;
    type: string | undefined;
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
            : STRING_LIKES.has(type ?? "")
              ? createLikeExpression(dimension, `%${searchText}%`)
              : undefined,
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
