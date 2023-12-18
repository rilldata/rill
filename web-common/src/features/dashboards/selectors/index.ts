import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createQueryServiceColumnTimeRange,
  createQueryServiceMetricsViewToplist,
  RpcStatus,
  V1ColumnTimeRangeResponse,
  V1MetricsViewSpec,
  V1MetricsViewToplistResponse,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";
import { derived, Readable } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";

export const useMetaQuery = <T = V1MetricsViewSpec>(
  ctx: StateManagers,
  selector?: (meta: V1MetricsViewSpec) => T
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
        ctx.queryClient
      ).subscribe(set);
    }
  );
};

export const useModelHasTimeSeries = (ctx: StateManagers) =>
  useMetaQuery(
    ctx,
    (meta) => !!meta?.timeDimension
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
  }
): Readable<QueryObserverResult<V1MetricsViewToplistResponse, RpcStatus>> => {
  return derived(
    [
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      ctx.metricsViewName,
      ctx.runtime,
    ],
    ([metricsExplorer, timeControls, metricViewName, runtime], set) => {
      return createQueryServiceMetricsViewToplist(
        runtime.instanceId,
        metricViewName,
        {
          dimensionName: dimension,
          measureNames: [metricsExplorer.leaderboardMeasureName],
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          limit: "100",
          offset: "0",
          sort: [],
          filter: {
            include: [
              {
                name: dimension,
                in: addNull ? [null] : [],
                like: [`%${searchText}%`],
              },
            ],
            exclude: [],
          },
        },
        {
          query: {
            queryClient: ctx.queryClient,
            enabled: timeControls.ready,
          },
        }
      ).subscribe(set);
    }
  );
};

export function createTimeRangeSummary(
  ctx: StateManagers
): CreateQueryResult<V1ColumnTimeRangeResponse> {
  return derived(
    [ctx.runtime, useMetaQuery(ctx)],
    ([runtime, metricsView], set) => {
      return createQueryServiceColumnTimeRange(
        runtime.instanceId,
        metricsView.data?.table,
        {
          columnName: metricsView.data?.timeDimension,
        },
        {
          query: {
            enabled: !!metricsView.data?.timeDimension,
            queryClient: ctx.queryClient,
          },
        }
      ).subscribe(set);
    }
  );
}
