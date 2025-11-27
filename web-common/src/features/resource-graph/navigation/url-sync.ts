/**
 * Composable for syncing resource graph state with URL parameters.
 *
 * This module extracts URL synchronization logic from ResourceGraph.svelte
 * to make it reusable and testable. It handles:
 * - Reading seeds from URL query parameters
 * - Writing expansion state to URL
 * - Cleaning up stale parameters
 * - Type-safe parameter handling
 *
 * URL API:
 * - `/graph?kind=metrics` - All MetricsView graphs
 * - `/graph?resource=orders` - Specific resource (defaults to MetricsView)
 * - `/graph?resource=model:orders` - Specific resource with explicit kind
 */

import { page } from "$app/stores";
import { goto } from "$app/navigation";
import { derived, type Readable } from "svelte/store";
import { debugLog } from "../shared/config";
import {
  URL_PARAMS,
  parseGraphUrlParams,
  urlParamsToSeeds,
} from "./seed-parser";

/**
 * Options for URL sync behavior.
 */
export interface GraphUrlSyncOptions {
  /**
   * Whether to sync expanded state with URL.
   * Set to false for embedded graphs that don't need URL persistence.
   */
  syncExpanded?: boolean;

  /**
   * Custom parameter names (for advanced use cases).
   */
  paramNames?: {
    seed?: string;
    expanded?: string;
  };

  /**
   * Callback when seeds change from URL.
   */
  onSeedsChange?: (seeds: string[]) => void;

  /**
   * Callback when expanded ID changes from URL.
   */
  onExpandedChange?: (id: string | null) => void;
}

/**
 * URL sync state.
 */
export interface GraphUrlSyncState {
  /**
   * Current seeds from URL.
   */
  seeds: Readable<string[]>;

  /**
   * Current expanded ID from URL (if syncExpanded is true).
   */
  expandedId: Readable<string | null>;

  /**
   * Set the expanded ID and update URL.
   */
  setExpanded: (id: string | null) => void;

  /**
   * Navigate to a specific seed configuration.
   */
  navigateToSeeds: (seeds: string[]) => void;

  /**
   * Add a seed to the current set.
   */
  addSeed: (seed: string) => void;

  /**
   * Remove a seed from the current set.
   */
  removeSeed: (seed: string) => void;

  /**
   * Clear all seeds.
   */
  clearSeeds: () => void;
}

/**
 * Create URL sync state for resource graph.
 *
 * @param options - Configuration options
 * @returns URL sync state and actions
 *
 * @example
 * // In a Svelte component
 * const urlSync = useGraphUrlSync({
 *   syncExpanded: true,
 *   onSeedsChange: (seeds) => console.log('Seeds changed:', seeds)
 * });
 *
 * // Access reactive state
 * $: console.log('Current seeds:', $urlSync.seeds);
 *
 * // Update URL
 * urlSync.setExpanded('model:orders');
 */
export function useGraphUrlSync(
  options: GraphUrlSyncOptions = {},
): GraphUrlSyncState {
  const {
    syncExpanded = true,
    paramNames = {},
    onSeedsChange,
    onExpandedChange,
  } = options;

  const expandedParam = paramNames.expanded ?? URL_PARAMS.EXPANDED;

  // Derive seeds from URL using new kind/resource parameters
  const seeds = derived(page, ($page) => {
    const params = parseGraphUrlParams($page.url);
    const filtered = urlParamsToSeeds(params);

    debugLog("URLSync", `Seeds from URL: ${filtered.length}`, filtered);

    return filtered;
  });

  // Derive expanded ID from URL (if enabled)
  const expandedId = derived(page, ($page) => {
    if (!syncExpanded) return null;

    const raw = $page.url.searchParams.get(expandedParam);
    const id = raw?.trim() || null;

    debugLog("URLSync", `Expanded ID from URL: ${id}`);

    return id;
  });

  // Track previous values to detect changes
  let prevSeeds: string[] = [];
  let prevExpandedId: string | null = null;

  // Subscribe to changes and call callbacks
  seeds.subscribe((value) => {
    if (JSON.stringify(value) !== JSON.stringify(prevSeeds)) {
      prevSeeds = value;
      onSeedsChange?.(value);
    }
  });

  expandedId.subscribe((value) => {
    if (value !== prevExpandedId) {
      prevExpandedId = value;
      onExpandedChange?.(value);
    }
  });

  /**
   * Update URL with new expanded ID.
   */
  function setExpanded(id: string | null): void {
    if (!syncExpanded) {
      debugLog("URLSync", "Skipping setExpanded: syncExpanded is false");
      return;
    }

    debugLog("URLSync", `Setting expanded: ${id}`);

    // Build new URL preserving other params
    const currentUrl = new URL(window.location.href);

    if (id) {
      currentUrl.searchParams.set(expandedParam, id);
    } else {
      currentUrl.searchParams.delete(expandedParam);
    }

    goto(currentUrl.pathname + currentUrl.search, {
      replaceState: true,
      noScroll: true,
      keepFocus: true,
    });
  }

  /**
   * Navigate to a specific seed configuration.
   * Uses the new URL API with kind/resource parameters.
   */
  function navigateToSeeds(newSeeds: string[]): void {
    debugLog("URLSync", `Navigating to seeds: ${newSeeds.length}`, newSeeds);

    const currentUrl = new URL(window.location.href);

    if (newSeeds.length === 0) {
      // Clear all resource and kind parameters
      currentUrl.searchParams.delete(URL_PARAMS.KIND);
      currentUrl.searchParams.delete(URL_PARAMS.RESOURCE);
      goto(currentUrl.pathname + currentUrl.search, {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
      return;
    }

    // Build URL with new API
    const params = new URLSearchParams();

    // Check if this is a kind-only filter
    const validKinds = ["metrics", "sources", "models", "dashboards"];
    if (
      newSeeds.length === 1 &&
      validKinds.includes(newSeeds[0].toLowerCase())
    ) {
      params.set(URL_PARAMS.KIND, newSeeds[0].toLowerCase());
    } else {
      // Use resource parameters for specific resources
      for (const seed of newSeeds) {
        if (seed && seed.trim()) {
          params.append(URL_PARAMS.RESOURCE, seed.trim());
        }
      }
    }

    // Preserve other query parameters (except old seed param for migration)
    for (const [key, value] of currentUrl.searchParams) {
      if (
        key !== URL_PARAMS.KIND &&
        key !== URL_PARAMS.RESOURCE &&
        key !== expandedParam &&
        key !== "seed" // Remove old seed param during migration
      ) {
        params.append(key, value);
      }
    }

    const newUrl = `${currentUrl.pathname}?${params.toString()}`;
    goto(newUrl, { replaceState: true, noScroll: true, keepFocus: true });
  }

  /**
   * Add a seed to the current set.
   */
  function addSeed(seed: string): void {
    if (!seed || !seed.trim()) {
      debugLog("URLSync", "Skipping addSeed: empty seed");
      return;
    }

    const currentSeeds = prevSeeds;
    if (currentSeeds.includes(seed)) {
      debugLog("URLSync", `Skipping addSeed: seed already exists: ${seed}`);
      return;
    }

    navigateToSeeds([...currentSeeds, seed]);
  }

  /**
   * Remove a seed from the current set.
   */
  function removeSeed(seed: string): void {
    const currentSeeds = prevSeeds;
    const filtered = currentSeeds.filter((s) => s !== seed);

    if (filtered.length === currentSeeds.length) {
      debugLog("URLSync", `Skipping removeSeed: seed not found: ${seed}`);
      return;
    }

    navigateToSeeds(filtered);
  }

  /**
   * Clear all seeds.
   */
  function clearSeeds(): void {
    debugLog("URLSync", "Clearing all seeds");
    navigateToSeeds([]);
  }

  return {
    seeds,
    expandedId,
    setExpanded,
    navigateToSeeds,
    addSeed,
    removeSeed,
    clearSeeds,
  };
}

/**
 * Parse seed parameters from a URL.
 * Uses the new kind/resource URL API.
 * Useful for server-side rendering or testing.
 */
export function parseSeedsFromUrl(url: URL | string): string[] {
  const urlObj = typeof url === "string" ? new URL(url) : url;
  const params = parseGraphUrlParams(urlObj);
  return urlParamsToSeeds(params);
}

/**
 * Build a URL with seed parameters.
 * Uses the new kind/resource URL API.
 * Useful for constructing links.
 */
export function buildGraphUrl(
  baseUrl: string,
  seeds: string[],
  expandedId?: string | null,
): string {
  const url = new URL(baseUrl, window.location.origin);

  // Check if this is a kind-only filter
  const validKinds = ["metrics", "sources", "models", "dashboards"];
  if (seeds.length === 1 && validKinds.includes(seeds[0].toLowerCase())) {
    url.searchParams.set(URL_PARAMS.KIND, seeds[0].toLowerCase());
  } else {
    for (const seed of seeds) {
      if (seed && seed.trim()) {
        url.searchParams.append(URL_PARAMS.RESOURCE, seed.trim());
      }
    }
  }

  if (expandedId) {
    url.searchParams.set(URL_PARAMS.EXPANDED, expandedId);
  }

  return url.pathname + url.search;
}
