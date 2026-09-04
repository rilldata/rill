import { writable, type Writable } from "svelte/store";

/**
 * Session-scoped test bindings for the component workspace preview, keyed by file path.
 * These are dev-time values used to render the preview; they are never persisted to
 * the component's YAML.
 */
const stores = new Map<string, Writable<Record<string, unknown>>>();

export function getPreviewArgsStore(
  filePath: string,
): Writable<Record<string, unknown>> {
  let store = stores.get(filePath);
  if (!store) {
    store = writable({});
    stores.set(filePath, store);
  }
  return store;
}
