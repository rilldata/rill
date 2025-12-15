import { EmbedStore } from "./embed-store";
import { getEmbedThemeStoreInstance } from "./embed-theme-store";
import { get } from "svelte/store";

/**
 * Resolves the theme name for embeds with priority:
 * 1. API value (including explicit null)
 * 2. Embed theme store value
 * 3. Initial embed URL theme
 */
export function resolveEmbedTheme(
  embedThemeValue?: string | null,
): string | null {
  // 1. Explicit API-provided value (including null to clear)
  if (embedThemeValue !== undefined) return embedThemeValue;

  // 2. Value from the scoped embed theme store, if non-null
  const embedThemeStore = getEmbedThemeStoreInstance();
  const storeValue = get(embedThemeStore);
  if (storeValue != null) return storeValue;

  // 3. Fallback to initial theme from the embed URL (captured in EmbedStore)
  const embedStore = EmbedStore.getInstance();
  return embedStore?.theme ?? null;
}
