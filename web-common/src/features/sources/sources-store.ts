import { writable, type Writable } from "svelte/store";

export enum DuplicateActions {
  None = "NONE",
  KeepBoth = "KEEP_BOTH",
  Overwrite = "OVERWRITE",
  Cancel = "CANCEL",
}

export const duplicateSourceAction: Writable<DuplicateActions> = writable(
  DuplicateActions.None,
);

export const duplicateSourceName: Writable<string | null> = writable(null);

export const sourceImportedPath = writable<string | null>(null);

/**
 * Tracks file paths that are pending source imports.
 * The UI registers paths here before writing source files,
 * so the watcher can identify new sources without relying on
 * the `definedAsSource` flag.
 */
const _pendingSourceImports = new Set<string>();

export const pendingSourceImports = {
  add(filePath: string) {
    _pendingSourceImports.add(filePath);
  },
  has(filePath: string) {
    return _pendingSourceImports.has(filePath);
  },
  delete(filePath: string) {
    _pendingSourceImports.delete(filePath);
  },
};
