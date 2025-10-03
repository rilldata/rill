import {
  type V1ListFilesResponse,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetInstance,
  runtimeServiceListFiles,
  runtimeServiceUnpackEmpty,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { EMPTY_PROJECT_TITLE } from "./constants";

export async function isProjectInitialized(instanceId: string) {
  try {
    const files = await queryClient.fetchQuery<V1ListFilesResponse>({
      queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
      queryFn: ({ signal }) => {
        return runtimeServiceListFiles(instanceId, undefined, signal);
      },
    });

    // Return true if `rill.yaml` exists, else false
    return !!files.files?.some(({ path }) => path === "/rill.yaml");
  } catch {
    return false;
  }
}

export async function handleUninitializedProject(instanceId: string) {
  // If the project is not initialized, determine what page to route to dependent on the OLAP connector
  const instance = await runtimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  const olapConnector = instance.instance?.olapConnector;

  if (!olapConnector) {
    throw new Error("OLAP connector is not defined");
  }

  // DuckDB-backed projects should head to the Welcome page for user-guided initialization
  if (olapConnector !== "duckdb") {
    // Clickhouse and Druid-backed projects should be initialized immediately
    await runtimeServiceUnpackEmpty(instanceId, {
      displayName: EMPTY_PROJECT_TITLE,
      olap: olapConnector, // Use the instance's configured OLAP connector
      force: true,
    });

    return true;
  }

  return false;
}
