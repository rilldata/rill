import { writable, Writable } from "svelte/store";
import { getContext } from "svelte";

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

export function createSourceStore(yaml: string) {
  return writable<SourceStore>({ clientYAML: yaml });
}

export function useSourceStore(): Writable<SourceStore> {
  return getContext("rill:app:source") as Writable<SourceStore>;
}
