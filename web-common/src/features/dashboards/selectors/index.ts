import {
  createQueryServiceMetricsViewToplist,
  createRuntimeServiceGetCatalogEntry,
  RpcStatus,
  V1MetricsViewToplistResponse,
  type V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { StateManagers } from "../state-managers/state-managers";
import { derived, Readable } from "svelte/store";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";

export const useMetaQuery = <T = V1MetricsView>(
  ctx: StateManagers,
  selector?: (meta: V1MetricsView) => T
): Readable<QueryObserverResult<T | V1MetricsView, RpcStatus>> => {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricViewName], set) => {
      return createRuntimeServiceGetCatalogEntry(
        runtime.instanceId,
        metricViewName,
        {
          query: {
            select: (data) =>
              selector
                ? selector(data?.entry?.metricsView)
                : data?.entry?.metricsView,
            queryClient: ctx.queryClient,
          },
        }
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
  const hasTimeSeriesStore = useModelHasTimeSeries(ctx);
  return derived(
    [ctx.dashboardStore, hasTimeSeriesStore, ctx.metricsViewName, ctx.runtime],
    ([metricsExplorer, hasTimeSeries, metricViewName, runtime], set) => {
      let topListParams = {
        dimensionName: dimension,
        measureNames: [metricsExplorer.leaderboardMeasureName],
        limit: "15",
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
      };
      if (hasTimeSeries) {
        topListParams = {
          ...topListParams,
          ...{
            timeStart: metricsExplorer?.selectedTimeRange?.start,
            timeEnd: metricsExplorer?.selectedTimeRange?.end,
          },
        };
      }

      return createQueryServiceMetricsViewToplist(
        runtime.instanceId,
        metricViewName,
        topListParams,
        {
          query: {
            queryClient: ctx.queryClient,
          },
        }
      ).subscribe(set);
    }
  );
};
