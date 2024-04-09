import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetFile,
  runtimeServiceGetResource,
} from "@rilldata/web-common/runtime-client";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import type { QueryFunction } from "@tanstack/svelte-query";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error } from "@sveltejs/kit";
import {
  ResourceKind,
  SingletonProjectParserName,
} from "@rilldata/web-common/features/entity-management/resource-selectors.js";

export async function load({ params, parent, depends }) {
  const { instanceId } = await parent();
  const dashboardName = params.name;

  const path = getFileAPIPathFromNameAndType(
    dashboardName,
    EntityType.Dashboard,
  );

  depends(path);

  const parser = fetchParser(instanceId);
  const fileData = fetchFile(instanceId, path);

  try {
    const [file, projectParser] = await Promise.all([fileData, parser]);
    const error =
      projectParser.resource?.projectParser?.state?.parseErrors?.find(
        (e) => e.filePath === "/" + path,
      );

    return {
      file: { ...file, path, error },
      dashboardName,
    };
  } catch (e) {
    throw error(404, "Dashboard not found");
  }
}

function fetchFile(instanceId: string, path: string) {
  const queryKey = getRuntimeServiceGetFileQueryKey(instanceId, path);

  const fileQuery: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetFile>>
  > = ({ signal }) => runtimeServiceGetFile(instanceId, path, signal);

  return queryClient.fetchQuery({
    queryKey,
    queryFn: fileQuery,
  });
}

function fetchParser(instanceId: string) {
  const parserParams = {
    "name.kind": ResourceKind.ProjectParser,
    "name.name": SingletonProjectParserName,
  };

  const projectParserQuery: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetResource>>
  > = ({ signal }) =>
    runtimeServiceGetResource(instanceId, parserParams, signal);

  return queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, parserParams),
    queryFn: projectParserQuery,
  });
}
