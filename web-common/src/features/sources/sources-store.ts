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

export interface SourceStore {
  clientYAML: string;
}

export const sourceStore: Writable<SourceStore> = writable({
  clientYAML: "",
});

export function useSourceStore(): Writable<SourceStore> {
  return sourceStore;
}
