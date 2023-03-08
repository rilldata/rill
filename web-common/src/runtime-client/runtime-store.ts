import { writable } from "svelte/store";

export interface Runtime {
  host: string;
  instanceId: string;
  jwt?: string;
}

export const runtime = writable<Runtime>();
