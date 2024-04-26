// This file checks for the existence of a `rill.yaml` file and handles the corresponding scenarios:
// - If the file exists, the app continues as normal.
// - If the file does not exist, the user is redirected to the Welcome page (DuckDB projects) or the project is initialized immediately (Clickhouse and Druid projects).

import { dev } from "$app/environment";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetInstance,
  runtimeServiceListFiles,
  type V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { redirect } from "@sveltejs/kit";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants";
import type { PageLoad } from "./$types";

export const ssr = false;

// When testing, we need to use the relative path to the server
const HOST = dev ? "http://localhost:9009" : "";
const INSTANCE_ID = "default";

const runtimeInit = {
  host: HOST,
  instanceId: INSTANCE_ID,
};

let init = false;

export const load: PageLoad = async ({ url, depends }) => {
  if (!init) runtime.set(runtimeInit);
  init = true;

  depends("init");

  const instanceId = runtimeInit.instanceId;

  const rillYAMLFiles = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: () => {
      return runtimeServiceListFiles(instanceId, undefined);
    },
  });

  const isProjectInitialized = Boolean(
    rillYAMLFiles.files?.some((file) => file.path === "/rill.yaml"),
  );

  if (!isProjectInitialized && url.pathname !== "/welcome") {
    await handleUninitializedProject(instanceId);
  } else if (isProjectInitialized && url.pathname === "/welcome") {
    throw redirect(303, "/");
  }

  return {
    ...runtimeInit,
    isProjectInitialized,
  };
};

import { runtimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";

async function handleUninitializedProject(instanceId: string) {
  // If the project is not initialized, determine what page to route to dependent on the OLAP connector
  const instance = await runtimeServiceGetInstance(instanceId);
  const olapConnector = instance.instance?.olapConnector;
  if (!olapConnector) {
    throw new Error("OLAP connector is not defined");
  }

  // DuckDB-backed projects should head to the Welcome page for user-guided initialization
  if (olapConnector === "duckdb") {
    throw redirect(303, "/welcome");
  }
  // Clickhouse and Druid-backed projects should be initialized immediately
  await runtimeServiceUnpackEmpty(instanceId, {
    title: EMPTY_PROJECT_TITLE,
    force: true,
  });
}
