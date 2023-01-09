/**
 * The ApplicationStore contains the state of the general application.
 * It does not contain any of the state for the entities; instead, it contains information
 * about things like the active entity and the application status.
 */
import { writable } from "svelte/store";

export type RuntimeState = {
  instanceId: string;
};
export const runtimeStore = writable<RuntimeState>({
  instanceId: null,
});
