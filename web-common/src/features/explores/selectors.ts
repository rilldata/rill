import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  type CreateQueryOptions,
  type QueryClient,
} from "@tanstack/svelte-query";
import {
  createRuntimeServiceGetExplore,
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1MetricsViewSpec,
  getRuntimeServiceGetExploreQueryOptions,
} from "@rilldata/web-common/runtime-client";
import type { ConnectError } from "@connectrpc/connect";
import { derived, type Readable } from "svelte/store";

export const PollIntervalWhenExploreReconciling = 1000;
export const PollIntervalWhenExploreErrored = 5000;

export function useExplore(
  client: RuntimeClient,
  exploreName: string,
  queryOptions?: Partial<
    CreateQueryOptions<V1GetExploreResponse, ConnectError, V1GetExploreResponse>
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
    CreateQueryOptions<
      V1GetExploreResponse,
      ConnectError,
      ExploreValidSpecResponse
    >
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

export function isExploreReconcilingForFirstTime(
  exploreResponse: V1GetExploreResponse,
) {
  if (!exploreResponse) return undefined;
  return (
    !exploreResponse.explore?.explore?.state?.validSpec &&
    !exploreResponse.explore?.meta?.reconcileError
  );
}

export function isExploreErrored(exploreResponse: V1GetExploreResponse) {
  if (!exploreResponse) return undefined;
  // Only consider errored when BOTH a reconcile error exists AND a validSpec
  // does not exist. If there's a validSpec (which can persist from a previous
  // spec), we serve that version of the dashboard to the user.
  return (
    !exploreResponse.explore?.explore?.state?.validSpec &&
    !!exploreResponse.explore?.meta?.reconcileError
  );
}
