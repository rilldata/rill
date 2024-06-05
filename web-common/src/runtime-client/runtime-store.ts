import { get, writable } from "svelte/store";
import { handleUninitializedProject } from "@rilldata/web-common/features/welcome/is-project-initialized";

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
    const onWelcomePage = window.location.pathname === "/welcome";
    const { instanceId } = get(runtime);
    if (!initialized) {
      const goToWelcomePage = await handleUninitializedProject(instanceId);
      if (goToWelcomePage && !onWelcomePage) {
        window.location.replace("/welcome");
      }
    } else if (onWelcomePage) {
      window.location.replace("/");
    }
    set(initialized);
  };

  return {
    subscribe,
    set,
    updateAndReroute,
  };
})();
