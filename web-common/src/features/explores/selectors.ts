import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  type CreateQueryOptions,
  type QueryFunction,
  type QueryClient,
} from "@tanstack/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetExplore,
  getRuntimeServiceGetExploreQueryKey,
  runtimeServiceGetExplore,
  getRuntimeServiceGetExploreQueryOptions,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import {
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import { derived, type Readable } from "svelte/store";

export function useExplore(
  client: RuntimeClient,
  exploreName: string,
  queryOptions?: Partial<
    CreateQueryOptions<V1GetExploreResponse, Error, V1GetExploreResponse>
  >,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetExplore(
    client,
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
  client: RuntimeClient,
  exploreName: string,
  queryOptions?: Partial<
    CreateQueryOptions<V1GetExploreResponse, Error, ExploreValidSpecResponse>
  >,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetExplore(
    client,
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
  client: RuntimeClient,
  exploreNameStore: Readable<string>,
) {
  return derived([exploreNameStore], ([exploreName]) =>
    getRuntimeServiceGetExploreQueryOptions(
      client,
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
  client: RuntimeClient,
  exploreName: string,
) {
  const queryParams = {
    name: exploreName,
  };
  const queryKey = getRuntimeServiceGetExploreQueryKey(
    client.instanceId,
    queryParams,
  );
  const queryFunction: QueryFunction<V1GetExploreResponse> = ({ signal }) =>
    runtimeServiceGetExplore(client, queryParams, { signal });

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
