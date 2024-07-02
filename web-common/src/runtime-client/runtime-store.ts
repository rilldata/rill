import { QueryClient } from "@tanstack/svelte-query";
import { writable } from "svelte/store";
import { invalidateRuntimeQueries } from "./invalidation";

export type AuthContext = "user" | "mock" | "magic" | "embed";

export interface JWT {
  token: string;
  // The time at which the JWT was received. We use this to calculate the JWT's expiration time.
  // We *could* parse the JWT to get the exact expiration time, but it's better to treat tokens as opaque.
  receivedAt: number;
  authContext: AuthContext;
}

export interface Runtime {
  host: string;
  instanceId: string;
  jwt?: JWT;
}

const createRuntimeStore = () => {
  const { subscribe, set, update } = writable<Runtime>({
    host: "",
    instanceId: "",
  });

  return {
    subscribe,
    update,
    set, // backwards-compatibility for web-local (where there's no JWT)
    setRuntime: async (
      queryClient: QueryClient,
      host: string,
      instanceId: string,
      jwt?: string,
      authContext?: AuthContext,
    ) => {
      if (jwt && !authContext) {
        throw new Error("authContext is required if jwt is provided");
      }

      let invalidate = false;

      update((current) => {
        // Invalidate the runtime if the auth context has changed
        // E.g. when switching from a normal user to a mocked user
        if (
          !!current.jwt?.authContext &&
          authContext !== current.jwt?.authContext
        ) {
          invalidate = true;
        }

        // Only update the store (particularly, the JWT `receivedAt`) if the values have changed
        if (
          host !== current.host ||
          instanceId !== current.instanceId ||
          jwt !== current.jwt?.token ||
          authContext !== current.jwt?.authContext
        ) {
          return {
            host,
            instanceId,
            jwt:
              jwt && authContext
                ? { token: jwt, receivedAt: Date.now(), authContext }
                : undefined,
          };
        }
        return current;
      });

      if (invalidate) await invalidateRuntimeQueries(queryClient, instanceId);
    },
  };
};

export const runtime = createRuntimeStore();
