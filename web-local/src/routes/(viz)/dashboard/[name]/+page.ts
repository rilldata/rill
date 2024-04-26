import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
} from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import type { QueryFunction } from "@tanstack/svelte-query";

export const load = async ({ parent, params, depends }) => {
  const { instanceId } = await parent();

  const dashboardName = params.name;

  depends(dashboardName, "dashboard");

  const queryParams = {
    "name.kind": ResourceKind.MetricsView,
    "name.name": dashboardName,
  };

  const queryKey = getRuntimeServiceGetResourceQueryKey(
    instanceId,
    queryParams,
  );

  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetResource>>
  > = ({ signal }) =>
    runtimeServiceGetResource(instanceId, queryParams, signal);

  try {
    const response = await queryClient.fetchQuery({
      queryFn: queryFunction,
      queryKey,
    });

    const metricsViewResource = response.resource;

    if (!metricsViewResource?.metricsView) {
      throw error(404, "Dashboard not found");
    }

    return {
      metricsView: metricsViewResource,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Dashboard not found");
  }
};
