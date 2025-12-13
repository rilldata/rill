import { EmbedStore } from "./embed-store";
import { getEmbedThemeStoreInstance } from "./embed-theme-store";
import { get } from "svelte/store";
import { page } from "$app/stores";

/**
 * Resolves the theme name with priority:
 * 1. embedThemeStore (set via API)
 * 2. EmbedStore.theme (from initial URL)
 * 3. URL params
 *
 * @param isEmbedded - Whether the dashboard is embedded
 * @param embedThemeValue - Current value from embedThemeStore (for reactive usage)
 * @param urlSearchParams - Optional URLSearchParams to check (defaults to current page URL)
 * @returns The resolved theme name, or null if no theme is set
 */
export function resolveEmbedTheme(
  isEmbedded: boolean,
  embedThemeValue?: string | null,
  urlSearchParams?: URLSearchParams,
): string | null {
  if (!isEmbedded) {
    const params = urlSearchParams ?? get(page).url.searchParams;
    return params.get("theme");
  }

  const embedThemeStore = getEmbedThemeStoreInstance();
  const apiTheme =
    embedThemeValue !== undefined ? embedThemeValue : get(embedThemeStore);
  if (apiTheme !== null) {
    return apiTheme;
  }

  const embedStore = EmbedStore.getInstance();
  const initialTheme = embedStore?.theme;
  if (initialTheme) {
    return initialTheme;
  }

  const params = urlSearchParams ?? get(page).url.searchParams;
  return params.get("theme");
}
