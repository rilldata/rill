import { queryClient } from "../../lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1GetResourceResponse,
} from "../../runtime-client";
import { ResourceKind, SingletonProjectParserName } from "./resource-selectors";

export function getProjectParserVersion(instanceId: string) {
  const projectParserQuery = queryClient.getQueryData<V1GetResourceResponse>(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    }),
  );

  if (!projectParserQuery?.resource?.meta?.version) {
    throw new Error("Project parser version not found");
  }

  return Number(projectParserQuery.resource.meta.version);
}

export async function waitForProjectParserVersion(
  instanceId: string,
  version: number,
) {
  const projectParserQuery = queryClient.getQueryData<V1GetResourceResponse>(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    }),
  );

  if (!projectParserQuery?.resource?.meta?.version) {
    throw new Error("Project parser version not found");
  }

  if (Number(projectParserQuery.resource.meta.version) >= version) {
    return;
  }

  await new Promise((resolve) => setTimeout(resolve, 300));

  return waitForProjectParserVersion(instanceId, version);
}
