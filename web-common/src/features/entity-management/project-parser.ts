import { queryClient } from "../../lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1GetResourceResponse,
} from "../../runtime-client";
import { ResourceKind, SingletonProjectParserName } from "./resource-selectors";

export function getProjectParserVersion(instanceId: string) {
  const projectParserQuery = queryClient.getQueryData<V1GetResourceResponse>(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
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
  let currentVersion = 0;

  while (currentVersion < version) {
    const projectParserQuery = queryClient.getQueryData<V1GetResourceResponse>(
      getRuntimeServiceGetResourceQueryKey(instanceId, {
        name: {
          kind: ResourceKind.ProjectParser,
          name: SingletonProjectParserName,
        },
      }),
    );

    if (!projectParserQuery?.resource?.meta?.version) {
      throw new Error("Project parser version not found");
    }

    currentVersion = Number(projectParserQuery.resource.meta.version);

    // If the current version is greater than or equal to the target version, we're done
    if (currentVersion >= version) return;

    // Wait before checking again
    await new Promise((resolve) => setTimeout(resolve, 300));
  }
}
