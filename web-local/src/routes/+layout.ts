import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { projectInitialized } from "@rilldata/web-common/runtime-client/runtime-store";
import { redirect } from "@sveltejs/kit";
import {
  runtimeServiceGetInstance,
  runtimeServiceUnpackEmpty,
} from "@rilldata/web-common/runtime-client/index.js";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants";
import { get } from "svelte/store";

export const ssr = false;

export async function load({ url }) {
  // Untrack url after upgrading to SvelteKit 2.0
  const onWelcomePage = url.pathname === "/welcome";

  const initialized = get(projectInitialized);

  if (!initialized && !onWelcomePage) {
    await handleUninitializedProject();
  } else if (initialized && onWelcomePage) {
    throw redirect(303, "/");
  }
}

async function handleUninitializedProject() {
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
  if (olapConnector === "duckdb") {
    throw redirect(303, "/welcome");
  } else {
    // Clickhouse and Druid-backed projects should be initialized immediately
    await runtimeServiceUnpackEmpty(instanceId, {
      title: EMPTY_PROJECT_TITLE,
      force: true,
    });
  }
}
