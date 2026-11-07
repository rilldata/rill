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

// Reads the parser version from the server rather than the cache:
// the cache lags behind if parser events were missed during a controller restart,
// and a stale starting version lets the wait below return before the new file is parsed.
export async function getProjectParserVersion(
  client: RuntimeClient,
  queryClient: QueryClient = globalQueryClient,
) {
  try {
    const parser = await fetchProjectParser(client, queryClient, true);
    return Number(parser?.meta?.version ?? 0);
  } catch (error) {
    // The project parser does not exist before the project is initialized.
    if (isNotFoundError(error)) return 0;
    if (!isControllerClosedError(error)) throw error;
    // The controller is restarting; fall back to the last known version.
    const cached = queryClient.getQueryData<V1GetResourceResponse>(
      getRuntimeServiceGetResourceQueryKey(client.instanceId, {
        name: {
          kind: ResourceKind.ProjectParser,
          name: SingletonProjectParserName,
        },
      }),
    );
    return Number(cached?.resource?.meta?.version ?? 0);
  }
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
