import {
  V1ListFilesResponse,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetInstance,
  runtimeServiceListFiles,
  runtimeServiceUnpackEmpty,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { EMPTY_PROJECT_TITLE } from "./constants";
import { get } from "svelte/store";

export async function isProjectInitialized(instanceId: string) {
  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(instanceId, undefined, signal);
    },
  });

  // Return true if `rill.yaml` exists, else false
  return files.files?.some((file) => file.path === "/rill.yaml");
}

export async function handleUninitializedProject() {
  const instanceId = get(runtime).instanceId;
  // If the project is not initialized, determine what page to route to dependent on the OLAP connector
  const instance = await runtimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  const olapConnector = instance.instance?.olapConnector;
  if (!olapConnector) {
    throw new Error("OLAP connector is not defined");
  }

  // DuckDB-backed projects should head to the Welcome page for user-guided initialization
  if (olapConnector === "duckdb" && window.location.pathname !== "/welcome") {
    window.location.replace("/welcome");
  } else {
    // Clickhouse and Druid-backed projects should be initialized immediately
    await runtimeServiceUnpackEmpty(instanceId, {
      title: EMPTY_PROJECT_TITLE,
      force: true,
    });
  }
}
