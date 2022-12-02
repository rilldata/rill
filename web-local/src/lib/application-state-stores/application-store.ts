/**
 * The ApplicationStore contains the state of the general application.
 * It does not contain any of the state for the entities; instead, it contains information
 * about things like the active entity and the application status.
 */
import { writable, Writable } from "svelte/store";

export enum DuplicateActions {
  None = "NONE",
  KeepBoth = "KEEP_BOTH",
  Overwrite = "OVERWRITE",
  Cancel = "CANCEL",
}

export const duplicateSourceAction: Writable<DuplicateActions> = writable(
  DuplicateActions.None
);

export const duplicateSourceName: Writable<string> = writable(null);

export type RuntimeState = {
  instanceId: string;
};
export const runtimeStore = writable<RuntimeState>({
  instanceId: null,
});
