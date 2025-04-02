import type { CreateQueryOptions, QueryFunction } from "@rilldata/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetExplore,
  getRuntimeServiceGetExploreQueryKey,
  runtimeServiceGetExplore,
  type RpcStatus,
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1MetricsViewSpec,
  getQueryServiceMetricsViewSchemaQueryKey,
  queryServiceMetricsViewSchema,
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

  return {
    explore: exploreResource,
    metricsView: metricsViewResource,
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
    queryFn: () => queryServiceMetricsViewSchema(instanceId, metricsViewName),
  });
  return schemaResp.schema ?? {};
}
