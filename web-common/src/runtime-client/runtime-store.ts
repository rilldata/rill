import { get, writable } from "svelte/store";
import { goto } from "$app/navigation";
import {
  runtimeServiceGetInstance,
  runtimeServiceUnpackEmpty,
} from "./gen/runtime-service/runtime-service";
import { EMPTY_PROJECT_TITLE } from "../features/welcome/constants";
import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized";

export interface JWT {
  token: string;
  // The time at which the JWT was received. We use this to calculate the JWT's expiration time.
  // We *could* parse the JWT to get the exact expiration time, but it's better to treat tokens as opaque.
  receivedAt: number;
}

export interface Runtime {
  host: string;
  instanceId: string;
  jwt?: JWT;
}

export const runtime = writable<Runtime>({
  host: "",
  instanceId: "",
});

export const projectInitialized = (() => {
  const { subscribe, set } = writable<boolean>(false);

  const updateAndReroute = async (initialized: boolean) => {
    if (!initialized) {
      await handleUninitializedProject();
    } else if (window.location.pathname === "/welcome") {
      await goto("/");
    }

    set(initialized);
  };

  return {
    subscribe,
    set: updateAndReroute,
    init: async () => {
      const instanceId = get(runtime).instanceId;
      const initialized = await isProjectInitialized(instanceId);
      await updateAndReroute(!!initialized);
    },
  };
})();

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
    await goto("/welcome");
  } else {
    // Clickhouse and Druid-backed projects should be initialized immediately
    await runtimeServiceUnpackEmpty(instanceId, {
      title: EMPTY_PROJECT_TITLE,
      force: true,
    });
  }
}
