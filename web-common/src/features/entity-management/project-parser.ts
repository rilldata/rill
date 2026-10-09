import { queryClient as globalQueryClient } from "../../lib/svelte-query/globalQueryClient";
import type { QueryClient } from "@tanstack/svelte-query";
import type { RuntimeClient } from "../../runtime-client/v2";
import { isControllerClosedError, isNotFoundError } from "../../lib/errors";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1GetResourceResponse,
} from "../../runtime-client";
import {
  fetchProjectParser,
  ResourceKind,
  SingletonProjectParserName,
} from "./resource-selectors";

export function getProjectParserVersion(
  instanceId: string,
  queryClient: QueryClient = globalQueryClient,
) {
  const projectParserQuery = queryClient.getQueryData<V1GetResourceResponse>(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    }),
  );

  if (!projectParserQuery?.resource?.meta?.version) {
    // Project parser is not present during the init. So dont throw error here, but assume version 0.
    return 0;
  }

  return Number(projectParserQuery.resource.meta.version);
}

export async function waitForProjectParserVersion(
  client: RuntimeClient,
  version: number,
  queryClient: QueryClient = globalQueryClient,
) {
  const deadline = Date.now() + 20_000;
  while (Date.now() < deadline) {
    try {
      // A controller restart can drop the SSE notification. Fetch the current
      // parser state instead of waiting indefinitely for the cache to change.
      const parser = await fetchProjectParser(client, queryClient, true);
      if (Number(parser?.meta?.version ?? 0) >= version) return;
    } catch (error) {
      if (!isControllerClosedError(error) && !isNotFoundError(error)) {
        throw error;
      }
    }
    await new Promise((resolve) => setTimeout(resolve, 300));
  }

  throw new Error(`Timed out waiting for project parser version ${version}`);
}
