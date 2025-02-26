import type { CreateQueryOptions, QueryFunction } from "@rilldata/svelte-query";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  convertPresetToExploreState,
  convertURLToExploreState,
} from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/url-state/getExploreStateFromSessionStorage";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { asyncWait, waitUntil } from "@rilldata/web-common/lib/waitUtils";
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
  type V1ExplorePreset,
  getQueryServiceMetricsViewSchemaQueryKey,
  queryServiceMetricsViewSchema,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error } from "@sveltejs/kit";
import { get } from "svelte/store";

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
) {
  console.log("fetchExploreSpec");
  await waitUntil(() => get(runtime).instanceId === instanceId, 5_000, 50);

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
    cacheTime: Infinity,
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

  const schema = await fetchMetricsViewSchema(
    instanceId,
    exploreSpec.metricsView ?? "",
  );

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
  const { partialExploreState: exploreStateFromYAMLConfig, errors } =
    convertPresetToExploreState(
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );

  return {
    explore: exploreResource,
    metricsView: metricsViewResource,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
    schema,
    errors,
  };
}

export async function fetchMetricsViewSchema(
  instanceId: string,
  metricsViewName: string,
) {
  const schemaResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewSchemaQueryKey(
      instanceId,
      metricsViewName,
    ),
    queryFn: async () => {
      await asyncWait(1_000);
      return queryServiceMetricsViewSchema(instanceId, metricsViewName);
    },
  });
  return schemaResp.schema ?? {};
}

export function getExploreStates(
  exploreName: string,
  prefix: string | undefined,
  searchParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec | undefined,
  exploreSpec: V1ExploreSpec | undefined,
  defaultExplorePreset: V1ExplorePreset,
) {
  if (!metricsViewSpec || !exploreSpec) {
    return {
      partialExploreStateFromUrl: <Partial<MetricsExplorerEntity>>{},
      exploreStateFromSessionStorage: undefined,
      errors: [],
    };
  }

  const { partialExploreState: partialExploreStateFromUrl, errors } =
    convertURLToExploreState(
      searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );

  const { exploreStateFromSessionStorage, errors: errorsFromLoad } =
    getExploreStateFromSessionStorage(
      exploreName,
      prefix,
      searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
  errors.push(...errorsFromLoad);

  return {
    partialExploreStateFromUrl,
    exploreStateFromSessionStorage,
    errors,
  };
}
