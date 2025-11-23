/**
 * Composable for syncing resource graph state with URL parameters.
 *
 * This module extracts URL synchronization logic from ResourceGraph.svelte
 * to make it reusable and testable. It handles:
 * - Reading seeds from URL query parameters
 * - Writing expansion state to URL
 * - Cleaning up stale parameters
 * - Type-safe parameter handling
 */

import { page } from "$app/stores";
import { goto } from "$app/navigation";
import { derived, type Readable } from "svelte/store";
import { debugLog } from "./graph-config";

/**
 * URL parameter names used for graph state.
 */
const PARAMS = {
  SEED: "seed",
  EXPANDED: "expanded",
} as const;

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

  const seedParam = paramNames.seed ?? PARAMS.SEED;
  const expandedParam = paramNames.expanded ?? PARAMS.EXPANDED;

  // Derive seeds from URL
  const seeds = derived(page, ($page) => {
    const rawSeeds = $page.url.searchParams.getAll(seedParam);
    const filtered = rawSeeds.filter((s) => s && s.trim());

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
   */
  function navigateToSeeds(newSeeds: string[]): void {
    debugLog("URLSync", `Navigating to seeds: ${newSeeds.length}`, newSeeds);

    const currentUrl = new URL(window.location.href);

    if (newSeeds.length === 0) {
      // Clear all seed parameters
      currentUrl.searchParams.delete(seedParam);
      goto(currentUrl.pathname + currentUrl.search, {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
      return;
    }

    // Build URL with multiple seed parameters
    const params = new URLSearchParams();

    for (const seed of newSeeds) {
      if (seed && seed.trim()) {
        params.append(seedParam, seed.trim());
      }
    }

    // Preserve other query parameters
    for (const [key, value] of currentUrl.searchParams) {
      if (key !== seedParam && key !== expandedParam) {
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
 * Useful for server-side rendering or testing.
 */
export function parseSeedsFromUrl(url: URL | string): string[] {
  const urlObj = typeof url === "string" ? new URL(url) : url;
  const rawSeeds = urlObj.searchParams.getAll(PARAMS.SEED);
  return rawSeeds.filter((s) => s && s.trim());
}

/**
 * Build a URL with seed parameters.
 * Useful for constructing links.
 */
export function buildGraphUrl(
  baseUrl: string,
  seeds: string[],
  expandedId?: string | null,
): string {
  const url = new URL(baseUrl, window.location.origin);

  for (const seed of seeds) {
    if (seed && seed.trim()) {
      url.searchParams.append(PARAMS.SEED, seed.trim());
    }
  }

  if (expandedId) {
    url.searchParams.set(PARAMS.EXPANDED, expandedId);
  }

  return url.pathname + url.search;
}
