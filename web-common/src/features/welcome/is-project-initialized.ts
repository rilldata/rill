import {
  type V1ListFilesResponse,
  V1ReconcileStatus,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetInstance,
  runtimeServiceListFiles,
  runtimeServiceUnpackEmpty,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceListResources } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import { EMPTY_PROJECT_TITLE } from "./constants";

const ProjectParserKind = "rill.runtime.v1.ProjectParser";

export async function isProjectInitialized(client: RuntimeClient) {
  try {
    const files = await queryClient.fetchQuery<V1ListFilesResponse>({
      queryKey: getRuntimeServiceListFilesQueryKey(client.instanceId, {}),
      queryFn: ({ signal }) => {
        return runtimeServiceListFiles(client, {}, { signal });
      },
    });

    // Return true if `rill.yaml` exists, else false
    return !!files.files?.some(({ path }) => path === "/rill.yaml");
  } catch {
    return false;
  }
}

export async function handleUninitializedProject(client: RuntimeClient) {
  // If the project is not initialized, determine what page to route to dependent on the OLAP connector
  const instance = await runtimeServiceGetInstance(client, {
    sensitive: true,
  });
  const olapConnector = instance.instance?.olapConnector;

  if (!olapConnector) {
    throw new Error("OLAP connector is not defined");
  }

  // DuckDB-backed projects should head to the Welcome page for user-guided initialization
  if (olapConnector !== "duckdb") {
    // Clickhouse and Druid-backed projects should be initialized immediately
    await runtimeServiceUnpackEmpty(client, {
      displayName: EMPTY_PROJECT_TITLE,
      olap: olapConnector, // Use the instance's configured OLAP connector
      force: true,
    });

    // Wait for all resources to finish reconciling before declaring initialized
    await waitForReconciliation(client);

    return true;
  }

  return false;
}

/**
 * Polls the resources API until all non-ProjectParser resources are idle.
 * This prevents the UI from rendering before the runtime has finished
 * parsing and reconciling resources after project initialization.
 */
export async function waitForReconciliation(
  client: RuntimeClient,
  timeoutMs = 60_000,
) {
  const settled = await asyncWaitUntil(async () => {
    try {
      const resp = await runtimeServiceListResources(client, {});
      const resources = resp.resources ?? [];
      if (resources.length === 0) return false;

      const dataResources = resources.filter(
        (r) => r.meta?.name?.kind !== ProjectParserKind,
      );

      return dataResources.every(
        (r) =>
          r.meta?.reconcileStatus ===
          V1ReconcileStatus.RECONCILE_STATUS_IDLE,
      );
    } catch {
      return false;
    }
  }, timeoutMs);

  if (!settled) {
    console.warn("Project reconciliation timed out; proceeding anyway");
  }
}
