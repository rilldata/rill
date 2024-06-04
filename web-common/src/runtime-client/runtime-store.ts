import { get, writable } from "svelte/store";
import {
  isProjectInitialized,
  handleUninitializedProject,
} from "@rilldata/web-common/features/welcome/is-project-initialized";

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
      window.location.replace("/");
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
