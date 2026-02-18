import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { get } from "svelte/store";

/**
 * Param type definitions for URL filter sync.
 * - "string": a single string value; empty string removes the param
 * - "array": a comma-separated list; empty array removes the param
 * - "enum": a single value with a default; the default value removes the param
 */
export type FilterParamDef =
  | { key: string; type: "string" }
  | { key: string; type: "array" }
  | { key: string; type: "enum"; defaultValue: string };

export function parseArrayParam(raw: string | null): string[] {
  return raw ? raw.split(",").filter(Boolean) : [];
}

export function parseStringParam(raw: string | null): string {
  return raw ?? "";
}

/**
 * Creates a URL filter sync utility for two-way URL <-> state synchronization.
 * Handles the common pattern of:
 * - Tracking lastSyncedSearch to distinguish external (back/forward) from programmatic navigation
 * - Serializing filter values to URL params
 * - Calling goto with replaceState
 */
export function createUrlFilterSync(paramDefs: FilterParamDef[]) {
  let lastSyncedSearch = "";

  return {
    /** Initialize from the current page URL (call once on setup) */
    init(url: URL) {
      lastSyncedSearch = url.search;
    },

    /** Returns true if the URL changed externally (back/forward navigation) */
    hasExternalNavigation(url: URL): boolean {
      return url.search !== lastSyncedSearch;
    },

    /** Update lastSyncedSearch after reading external navigation changes */
    markSynced(url: URL) {
      lastSyncedSearch = url.search;
    },

    /**
     * Write filter values to URL params and navigate.
     * Reads current URL via get(page) to avoid making $page a reactive
     * dependency in calling $: blocks (which would cause an infinite goto loop).
     */
    syncToUrl(values: Record<string, string | string[]>) {
      const newUrl = new URL(get(page).url);

      for (const def of paramDefs) {
        const value = values[def.key];
        let serialized: string | null = null;

        if (def.type === "array") {
          const arr = value as string[];
          serialized = arr.length > 0 ? arr.join(",") : null;
        } else if (def.type === "enum") {
          serialized = value !== def.defaultValue ? (value as string) : null;
        } else {
          serialized = (value as string) || null;
        }

        if (serialized !== null) {
          newUrl.searchParams.set(def.key, serialized);
        } else {
          newUrl.searchParams.delete(def.key);
        }
      }

      lastSyncedSearch = newUrl.search;
      void goto(newUrl.pathname + newUrl.search, {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
    },
  };
}
