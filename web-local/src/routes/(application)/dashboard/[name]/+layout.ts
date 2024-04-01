import {
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetFile,
  runtimeServiceGetResource,
} from "@rilldata/web-common/runtime-client/index.js";
import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import type { QueryFunction } from "@tanstack/svelte-query";
import { error, redirect } from "@sveltejs/kit";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";

export const ssr = false;

export async function load({ parent, params }) {
  const parentData = await parent();
  const dashboardName = params.name;

  const filePath = getFilePathFromNameAndType(
    dashboardName,
    EntityType.MetricsDefinition,
  );

  const instanceId = parentData.instance?.instanceId ?? "default";

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetFile>>
  > = ({ signal }) => runtimeServiceGetFile(instanceId, filePath, signal);

  const resourceParams = {
    "name.kind": ResourceKind.MetricsView,
    "name.name": dashboardName,
  };

  const resourceFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetResource>>
  > = ({ signal }) =>
    runtimeServiceGetResource(instanceId, resourceParams, signal);

  const fileQuery = queryClient.fetchQuery({
    queryKey: [filePath],
    queryFn,
  });

  const resourceQuery = queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, resourceParams),
    queryFn: resourceFn,
  });

  try {
    const file = await fileQuery;
    const metricsViewResponse = await resourceQuery;

    if (!metricsViewResponse.resource?.metricsView?.state?.validSpec) {
      throw redirect(303, "/edit");
    }

    return {
      file,
      metricsView: metricsViewResponse.resource,
      dimensions:
        metricsViewResponse.resource.metricsView.state.validSpec.dimensions ??
        [],
      measures:
        metricsViewResponse.resource.metricsView.state.validSpec.measures ?? [],
    };
  } catch {
    throw error(404, "dashboard not found");
  }
}
