/**
 * Centralized cache management for resource graph.
 *
 * This module provides a unified interface for managing all graph-related caching:
 * - Node positions (for stable layouts across renders)
 * - Group assignments (which resources belong to which graphs)
 * - Group labels (display names for graph groups)
 * - Resource references (dependency relationships)
 *
 * The cache is persisted to localStorage and automatically synced with in-memory Maps.
 */

import { localStorageStore } from "@rilldata/web-common/lib/store-utils/local-storage";
import {
  CACHE_NAMESPACE,
  CACHE_KEY_PATTERN,
  debugLog,
  PERFORMANCE_CONFIG,
} from "./graph-config";

/**
 * Structure of persisted cache data.
 */
export type PersistedCache = {
  positions: Record<string, { x: number; y: number }>;
  assignments: Record<string, string>; // resourceId -> groupId
  labels: Record<string, string>; // groupId -> label
  refs: Record<string, string[]>; // dependentId -> sourceIds[]
};

const DEFAULT_CACHE: PersistedCache = {
  positions: {},
  assignments: {},
  labels: {},
  refs: {},
};

/**
 * Centralized cache manager for resource graph data.
 *
 * This class encapsulates all caching logic and provides:
 * - Type-safe accessors for cached data
 * - Automatic localStorage persistence with debouncing
 * - Cache health diagnostics
 * - Explicit initialization for SSR compatibility
 * - Debug utilities
 */
export class GraphCacheManager {
  private initialized = false;
  private store = localStorageStore<PersistedCache>(
    CACHE_NAMESPACE,
    DEFAULT_CACHE,
  );

  // In-memory caches for fast access
  private positions = new Map<string, { x: number; y: number }>();
  private assignments = new Map<string, string>();
  private labels = new Map<string, string>();
  private refs = new Map<string, string[]>();

  // Track if we have pending changes to persist
  private dirty = false;
  private persistTimer: ReturnType<typeof setTimeout> | null = null;

  /**
   * Initialize the cache manager.
   * Call this explicitly on client-side mount to avoid SSR issues.
   */
  initialize(): void {
    if (this.initialized) {
      debugLog("Cache", "Already initialized, skipping");
      return;
    }

    if (typeof window === "undefined") {
      debugLog("Cache", "Skipping initialization in SSR environment");
      return;
    }

    debugLog("Cache", "Initializing cache manager");

    // Subscribe to store changes and sync to in-memory Maps
    this.store.subscribe((cache) => {
      this.syncFromStore(cache);
    });

    // Clean up orphaned caches from old versions
    this.cleanupOrphanedCaches();

    this.initialized = true;

    // Expose on window for debugging
    if (typeof window !== "undefined") {
      (window as any).__RESOURCE_GRAPH_CACHE = this;
    }

    debugLog("Cache", "Initialization complete");
  }

  /**
   * Check if cache manager is initialized.
   */
  isInitialized(): boolean {
    return this.initialized;
  }

  /**
   * Sync in-memory Maps from store data.
   */
  private syncFromStore(cache: PersistedCache): void {
    debugLog("Cache", "Syncing from store", {
      positions: Object.keys(cache.positions).length,
      assignments: Object.keys(cache.assignments).length,
      labels: Object.keys(cache.labels).length,
      refs: Object.keys(cache.refs).length,
    });

    // Update positions
    this.positions.clear();
    for (const [k, v] of Object.entries(cache.positions)) {
      this.positions.set(k, v);
    }

    // Update assignments
    this.assignments.clear();
    for (const [k, v] of Object.entries(cache.assignments)) {
      this.assignments.set(k, v);
    }

    // Update labels
    this.labels.clear();
    for (const [k, v] of Object.entries(cache.labels)) {
      this.labels.set(k, v);
    }

    // Update refs
    this.refs.clear();
    for (const [k, v] of Object.entries(cache.refs)) {
      this.refs.set(k, v);
    }

    this.dirty = false;
  }

  /**
   * Mark cache as dirty and schedule persistence.
   */
  private markDirty(): void {
    this.dirty = true;

    // Clear existing timer
    if (this.persistTimer) {
      clearTimeout(this.persistTimer);
    }

    // Schedule debounced persist
    this.persistTimer = setTimeout(() => {
      this.persist();
    }, PERFORMANCE_CONFIG.CACHE_WRITE_DEBOUNCE_MS);
  }

  /**
   * Immediately persist in-memory state to localStorage.
   */
  persist(): void {
    if (!this.initialized) {
      debugLog("Cache", "Cannot persist: not initialized");
      return;
    }

    if (!this.dirty) {
      debugLog("Cache", "Skipping persist: no changes");
      return;
    }

    debugLog("Cache", "Persisting cache to localStorage");

    this.store.set({
      positions: Object.fromEntries(this.positions),
      assignments: Object.fromEntries(this.assignments),
      labels: Object.fromEntries(this.labels),
      refs: Object.fromEntries(this.refs),
    });

    this.dirty = false;
  }

  /**
   * Clean up orphaned cache entries from old versions.
   */
  private cleanupOrphanedCaches(): void {
    try {
      if (typeof window === "undefined" || !window.localStorage) return;

      const keys = Object.keys(window.localStorage);
      let cleanedCount = 0;

      for (const key of keys) {
        if (CACHE_KEY_PATTERN.test(key) && key !== CACHE_NAMESPACE) {
          window.localStorage.removeItem(key);
          cleanedCount++;
        }
      }

      if (cleanedCount > 0) {
        debugLog(
          "Cache",
          `Cleaned up ${cleanedCount} orphaned cache ${cleanedCount === 1 ? "entry" : "entries"}`,
        );
      }
    } catch (error) {
      console.warn("[ResourceGraph] Failed to cleanup orphaned caches:", error);
    }
  }

  // Position accessors
  getPosition(key: string): { x: number; y: number } | undefined {
    return this.positions.get(key);
  }

  setPosition(key: string, position: { x: number; y: number }): void {
    this.positions.set(key, position);
    this.markDirty();
  }

  // Assignment accessors
  getAssignment(resourceId: string): string | undefined {
    return this.assignments.get(resourceId);
  }

  setAssignment(resourceId: string, groupId: string): void {
    this.assignments.set(resourceId, groupId);
    this.markDirty();
  }

  // Label accessors
  getLabel(groupId: string): string | undefined {
    return this.labels.get(groupId);
  }

  setLabel(groupId: string, label: string): void {
    this.labels.set(groupId, label);
    this.markDirty();
  }

  // Refs accessors
  getRefs(dependentId: string): string[] | undefined {
    return this.refs.get(dependentId);
  }

  setRefs(dependentId: string, sourceIds: string[]): void {
    this.refs.set(dependentId, sourceIds);
    this.markDirty();
  }

  addRef(dependentId: string, sourceId: string): void {
    const existing = this.refs.get(dependentId) ?? [];
    if (!existing.includes(sourceId)) {
      this.refs.set(dependentId, [...existing, sourceId]);
      this.markDirty();
    }
  }

  /**
   * Get cache health statistics for debugging.
   */
  getHealthStats(): {
    initialized: boolean;
    dirty: boolean;
    positions: number;
    assignments: number;
    labels: number;
    refs: number;
    totalEntries: number;
  } {
    return {
      initialized: this.initialized,
      dirty: this.dirty,
      positions: this.positions.size,
      assignments: this.assignments.size,
      labels: this.labels.size,
      refs: this.refs.size,
      totalEntries:
        this.positions.size +
        this.assignments.size +
        this.labels.size +
        this.refs.size,
    };
  }

  /**
   * Clear all cached data.
   * Useful for debugging or when cache becomes corrupted.
   */
  clearAll(): void {
    debugLog("Cache", "Clearing all cached data");

    this.positions.clear();
    this.assignments.clear();
    this.labels.clear();
    this.refs.clear();

    this.markDirty();
    this.persist();
  }

  /**
   * Export cache data for debugging.
   */
  export(): PersistedCache {
    return {
      positions: Object.fromEntries(this.positions),
      assignments: Object.fromEntries(this.assignments),
      labels: Object.fromEntries(this.labels),
      refs: Object.fromEntries(this.refs),
    };
  }

  /**
   * Import cache data (useful for testing or migrations).
   */
  import(data: PersistedCache): void {
    debugLog("Cache", "Importing cache data");

    this.positions.clear();
    this.assignments.clear();
    this.labels.clear();
    this.refs.clear();

    for (const [k, v] of Object.entries(data.positions)) {
      this.positions.set(k, v);
    }
    for (const [k, v] of Object.entries(data.assignments)) {
      this.assignments.set(k, v);
    }
    for (const [k, v] of Object.entries(data.labels)) {
      this.labels.set(k, v);
    }
    for (const [k, v] of Object.entries(data.refs)) {
      this.refs.set(k, v);
    }

    this.markDirty();
    this.persist();
  }

  /**
   * Cleanup resources (call on component unmount in tests).
   */
  destroy(): void {
    if (this.persistTimer) {
      clearTimeout(this.persistTimer);
      this.persistTimer = null;
    }

    // Force final persist
    if (this.dirty) {
      this.persist();
    }

    this.initialized = false;

    debugLog("Cache", "Cache manager destroyed");
  }
}

/**
 * Singleton instance of cache manager.
 */
export const graphCache = new GraphCacheManager();
