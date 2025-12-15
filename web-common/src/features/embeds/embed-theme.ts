import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage";
import { get } from "svelte/store";
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

/**
 * Resolves the theme name for embeds with priority:
 * 1. Value from the scoped embed theme store, if non-null
 * 2. Initial embed URL theme
 *
 * Note: This function is always called with the embed theme store value,
 * so we get it directly from the store rather than accepting it as a parameter.
 */
export function resolveEmbedTheme(): string | null {
  // 1. Value from the scoped embed theme store, if non-null
  const embedThemeStore = getEmbedThemeStoreInstance();
  const storeValue = get(embedThemeStore);
  if (storeValue != null) return storeValue;

  // 2. Fallback to initial theme from the embed URL (captured in EmbedStore)
  const embedStore = EmbedStore.getInstance();
  return embedStore?.theme ?? null;
}
