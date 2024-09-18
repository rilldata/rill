import type { CreateQueryOptions } from "@rilldata/svelte-query";
import {
  createRuntimeServiceGetExplore,
  RpcStatus,
  V1ExploreSpec,
  V1GetExploreResponse,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

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

export type ValidExploreResponse = {
  explore: V1ExploreSpec | undefined;
  metricsView: V1MetricsViewSpec | undefined;
};
export function useValidExplore(instanceId: string, exploreName: string) {
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: {
        select: (data) =>
          <ValidExploreResponse>{
            explore: data.explore?.explore?.state?.validSpec,
            metricsView: data.metricsView?.metricsView?.state?.validSpec,
          },
      },
    },
  );
}
