import { writable } from "svelte/store";

export const runtimeUrl = import.meta.env.DEV ? "http://localhost:9009" : "";

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
  host: runtimeUrl,
  instanceId: "default",
  jwt: undefined,
});
