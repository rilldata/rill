import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage";
import { EmbedStore } from "./embed-store";

function getEmbedThemeStoreKey(): string {
  const embedStore = EmbedStore.getInstance();
  const embedId = embedStore?.embedId ?? "default";
  return `rill:embed:theme:${embedId}`;
}

export function getEmbedThemeStore() {
  return sessionStorageStore<string | null>(getEmbedThemeStoreKey(), undefined);
}

let _embedThemeStore: ReturnType<
  typeof sessionStorageStore<string | null>
> | null = null;

export function getEmbedThemeStoreInstance() {
  if (!_embedThemeStore) {
    _embedThemeStore = getEmbedThemeStore();
  }
  return _embedThemeStore;
}

export function clearEmbedThemeStore() {
  if (typeof window !== "undefined" && window.sessionStorage) {
    const key = getEmbedThemeStoreKey();
    try {
      window.sessionStorage.removeItem(key);
    } catch {
      // Ignore errors
    }
  }
  _embedThemeStore = null;
}
