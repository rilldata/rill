import type { CreateQueryOptions, QueryFunction } from "@rilldata/svelte-query";
import { restorePersistedDashboardState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetExplore,
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  queryServiceMetricsViewTimeRange,
  runtimeServiceGetExplore,
  type RpcStatus,
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import { error } from "@sveltejs/kit";

export function useExplore(
  instanceId: string,
  exploreName: string,
  queryOptions?: CreateQueryOptions<
    V1GetExploreResponse,
    ErrorType<RpcStatus>,
    V1GetExploreResponse
  >,
) {
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: queryOptions,
    },
  );
}

export type ExploreValidSpecResponse = {
  explore: V1ExploreSpec | undefined;
  metricsView: V1MetricsViewSpec | undefined;
};
export function useExploreValidSpec(
  instanceId: string,
  exploreName: string,
  queryOptions?: CreateQueryOptions<
    V1GetExploreResponse,
    ErrorType<RpcStatus>,
    ExploreValidSpecResponse
  >,
) {
  const defaultQueryOptions: CreateQueryOptions<
    V1GetExploreResponse,
    ErrorType<RpcStatus>,
    ExploreValidSpecResponse
  > = {
    select: (data) =>
      <ExploreValidSpecResponse>{
        explore: data.explore?.explore?.state?.validSpec,
        metricsView: data.metricsView?.metricsView?.state?.validSpec,
      },
    queryClient,
    enabled: !!exploreName,
  };
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: {
        ...defaultQueryOptions,
        ...queryOptions,
      },
    },
  );
}

export async function fetchExploreSpec(
  instanceId: string,
  exploreName: string,
  prefix: string | undefined,
) {
  const queryParams = {
    name: exploreName,
  };
  const queryKey = getRuntimeServiceGetExploreQueryKey(instanceId, queryParams);
  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetExplore>>
  > = ({ signal }) => runtimeServiceGetExplore(instanceId, queryParams, signal);

  const response = await queryClient.fetchQuery({
    queryFn: queryFunction,
    queryKey,
    staleTime: Infinity,
  });

  const exploreResource = response.explore;
  const metricsViewResource = response.metricsView;

  if (!exploreResource?.explore) {
    throw error(404, "Explore not found");
  }
  if (!metricsViewResource?.metricsView) {
    throw error(404, "Metrics view not found");
  }

  const metricsViewSpec =
    metricsViewResource.metricsView.state?.validSpec ?? {};
  const exploreSpec = exploreResource.explore.state?.validSpec ?? {};

  let fullTimeRange: V1MetricsViewTimeRangeResponse | undefined = undefined;
  const metricsViewName = exploreSpec.metricsView;
  if (metricsViewSpec.timeDimension && metricsViewName) {
    fullTimeRange = await queryClient.fetchQuery({
      queryFn: () =>
        queryServiceMetricsViewTimeRange(instanceId, metricsViewName, {}),
      queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
        instanceId,
        metricsViewName,
        {},
      ),
      staleTime: Infinity,
      cacheTime: Infinity,
    });
  }

  const defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    fullTimeRange,
  );
  const { partialExploreState: initExploreState, errors } =
    convertPresetToExploreState(
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );

  let initLoadedOutsideOfURL = false;
  const stateFromLocalStorage = restorePersistedDashboardState(
    exploreSpec,
    (prefix ?? "") + exploreName,
  );
  if (stateFromLocalStorage) {
    initLoadedOutsideOfURL = true;
    Object.assign(initExploreState, stateFromLocalStorage);
  }

  return {
    explore: exploreResource,
    metricsView: metricsViewResource,
    defaultExplorePreset,
    initExploreState,
    initLoadedOutsideOfURL,
    errors,
  };
}
