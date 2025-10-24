import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import type {
  CreateQueryOptions,
  QueryFunction,
  QueryClient,
} from "@tanstack/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetExplore,
  getRuntimeServiceGetExploreQueryKey,
  runtimeServiceGetExplore,
  type RpcStatus,
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1MetricsViewSpec,
  getRuntimeServiceGetExploreQueryOptions,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import { error } from "@sveltejs/kit";
import { derived, type Readable } from "svelte/store";

export function useExplore(
  instanceId: string,
  exploreName: string,
  queryOptions?: Partial<
    CreateQueryOptions<
      V1GetExploreResponse,
      ErrorType<RpcStatus>,
      V1GetExploreResponse
    >
  >,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: queryOptions,
    },
    queryClient,
  );
}

export type ExploreValidSpecResponse = {
  explore: V1ExploreSpec | undefined;
  metricsView: V1MetricsViewSpec | undefined;
};
export function useExploreValidSpec(
  instanceId: string,
  exploreName: string,
  queryOptions?: Partial<
    CreateQueryOptions<
      V1GetExploreResponse,
      ErrorType<RpcStatus>,
      ExploreValidSpecResponse
    >
  >,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: {
        select: (data) =>
          <ExploreValidSpecResponse>{
            explore: data.explore?.explore?.state?.validSpec,
            metricsView: data.metricsView?.metricsView?.state?.validSpec,
          },

        enabled: !!exploreName,
        ...queryOptions,
      },
    },
    queryClient,
  );
}

export function getExploreValidSpecQueryOptions(
  exploreNameStore: Readable<string>,
) {
  return derived([runtime, exploreNameStore], ([{ instanceId }, exploreName]) =>
    getRuntimeServiceGetExploreQueryOptions(
      instanceId,
      {
        name: exploreName,
      },
      {
        query: {
          select: (data) => ({
            exploreSpec: data.explore?.explore?.state?.validSpec,
            metricsViewSpec: data.metricsView?.metricsView?.state?.validSpec,
          }),
          enabled: !!exploreName,
        },
      },
    ),
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
