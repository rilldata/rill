import { writable } from "svelte/store";

export type Runtime = {
  host: string;
  instanceId: string;
};

export const runtime = writable<Runtime>();
