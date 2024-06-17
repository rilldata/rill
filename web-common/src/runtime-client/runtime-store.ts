import { writable } from "svelte/store";

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

const createRuntimeStore = () => {
  const { subscribe, set, update } = writable<Runtime>({
    host: "",
    instanceId: "",
  });

  return {
    subscribe,
    set, // backwards-compatibility for web-local (where there's no JWT)
    setRuntime: (host: string, instanceId: string, jwt?: string) => {
      update((current) => {
        // Only update the store (particularly, the JWT `receivedAt`) if the values have changed
        if (
          host !== current.host ||
          instanceId !== current.instanceId ||
          jwt !== current.jwt?.token
        ) {
          return {
            host,
            instanceId,
            jwt: jwt ? { token: jwt, receivedAt: Date.now() } : undefined,
          };
        }
        return current;
      });
    },
  };
};

export const runtime = createRuntimeStore();
