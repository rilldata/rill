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

export const duplicateSourceName: Writable<string | null> = writable(null);

interface SourceStore {
  clientYAML: string | null;
}

// Dictionary of source stores
const sourceStores: { [key: string]: Writable<SourceStore> } = {};

function createSourceStore(): Writable<SourceStore> {
  return writable({ clientYAML: null });
}

export function useSourceStore(sourceName: string): Writable<SourceStore> {
  if (!sourceStores[sourceName]) {
    sourceStores[sourceName] = createSourceStore();
  }

  return sourceStores[sourceName];
}

export const sourceImportedName = writable<string | null>(null);
