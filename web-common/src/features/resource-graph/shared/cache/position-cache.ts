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
  CACHE_CONFIG,
} from "../config";

/**
 * Structure of persisted cache data.
 */
export type PersistedCache = {
  positions: Record<string, { x: number; y: number }>;
  assignments: Record<string, string>; // resourceId -> groupId
  labels: Record<string, string>; // groupId -> label
};

const DEFAULT_CACHE: PersistedCache = {
  positions: {},
  assignments: {},
  labels: {},
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

  // Track if we have pending changes to persist
  private dirty = false;
  private persistTimer: ReturnType<typeof setTimeout> | null = null;

  // Track cache health
  private quotaExceeded = false;
  private lastPruneTime = 0;

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
   * Includes error handling for quota exceeded, disabled storage, and size limits.
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

    // Skip if quota was previously exceeded (until cleared)
    if (this.quotaExceeded) {
      debugLog(
        "Cache",
        "Skipping persist: quota exceeded, cache disabled until cleared",
      );
      return;
    }

    debugLog("Cache", "Persisting cache to localStorage");

    try {
      const data: PersistedCache = {
        positions: Object.fromEntries(this.positions),
        assignments: Object.fromEntries(this.assignments),
        labels: Object.fromEntries(this.labels),
      };

      // Check size before writing
      const dataSize = this.estimateCacheSize(data);
      debugLog("Cache", `Estimated cache size: ${dataSize} bytes`);

      if (dataSize > CACHE_CONFIG.MAX_SIZE_BYTES) {
        console.warn(
          `[ResourceGraph] Cache size (${dataSize} bytes) exceeds limit (${CACHE_CONFIG.MAX_SIZE_BYTES} bytes), pruning...`,
        );
        this.pruneOldestEntries();
        // Retry persist after pruning
        return this.persist();
      }

      // Attempt to write to localStorage
      this.store.set(data);
      this.dirty = false;
      this.quotaExceeded = false; // Reset quota flag on success
      debugLog("Cache", "Persist successful");
    } catch (error) {
      if (
        error instanceof DOMException &&
        (error.name === "QuotaExceededError" ||
          error.name === "NS_ERROR_DOM_QUOTA_REACHED")
      ) {
        console.warn(
          "[ResourceGraph] LocalStorage quota exceeded, clearing cache and disabling until manually cleared",
        );
        this.quotaExceeded = true;
        this.clearAll();
      } else if (
        error instanceof DOMException &&
        error.name === "SecurityError"
      ) {
        console.warn(
          "[ResourceGraph] LocalStorage access denied (private browsing or disabled), cache will not persist",
        );
        this.quotaExceeded = true; // Disable further writes
      } else {
        console.error("[ResourceGraph] Failed to persist cache:", error);
        // Don't set quotaExceeded for unknown errors, allow retry
      }
      this.dirty = false; // Clear dirty flag to prevent infinite retry
    }
  }

  /**
   * Estimate cache size in bytes using JSON serialization.
   * This is approximate but sufficient for quota management.
   */
  private estimateCacheSize(data: PersistedCache): number {
    try {
      return JSON.stringify(data).length * 2; // UTF-16 uses 2 bytes per char
    } catch (error) {
      console.error("[ResourceGraph] Failed to estimate cache size:", error);
      return 0;
    }
  }

  /**
   * Prune oldest entries based on LRU strategy.
   * Removes oldest positions first, then assignments.
   * Labels are kept as they're small and important for recovery.
   */
  private pruneOldestEntries(): void {
    const now = Date.now();
    // Prevent excessive pruning (max once per 5 seconds)
    if (now - this.lastPruneTime < CACHE_CONFIG.MIN_PRUNE_INTERVAL_MS) {
      debugLog("Cache", "Skipping prune: too soon since last prune");
      return;
    }

    this.lastPruneTime = now;
    debugLog("Cache", "Pruning oldest cache entries");

    const initialSize =
      this.positions.size + this.assignments.size + this.labels.size;

    // Prune 25% of positions (oldest positions are least likely to be reused)
    const positionsToRemove = Math.ceil(this.positions.size * 0.25);
    const positionKeys = Array.from(this.positions.keys());
    for (let i = 0; i < positionsToRemove && i < positionKeys.length; i++) {
      this.positions.delete(positionKeys[i]);
    }

    // Prune 25% of assignments if positions pruning wasn't enough
    if (positionsToRemove < 10) {
      const assignmentsToRemove = Math.ceil(this.assignments.size * 0.25);
      const assignmentKeys = Array.from(this.assignments.keys());
      for (
        let i = 0;
        i < assignmentsToRemove && i < assignmentKeys.length;
        i++
      ) {
        this.assignments.delete(assignmentKeys[i]);
      }
    }

    const finalSize =
      this.positions.size + this.assignments.size + this.labels.size;

    debugLog(
      "Cache",
      `Pruned ${initialSize - finalSize} entries (${initialSize} → ${finalSize})`,
    );

    this.markDirty();
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

  /**
   * Get cache health statistics for debugging.
   */
  getHealthStats(): {
    initialized: boolean;
    dirty: boolean;
    quotaExceeded: boolean;
    positions: number;
    assignments: number;
    labels: number;
    totalEntries: number;
    estimatedSizeBytes: number;
  } {
    const data = {
      positions: Object.fromEntries(this.positions),
      assignments: Object.fromEntries(this.assignments),
      labels: Object.fromEntries(this.labels),
    };

    return {
      initialized: this.initialized,
      dirty: this.dirty,
      quotaExceeded: this.quotaExceeded,
      positions: this.positions.size,
      assignments: this.assignments.size,
      labels: this.labels.size,
      totalEntries:
        this.positions.size + this.assignments.size + this.labels.size,
      estimatedSizeBytes: this.estimateCacheSize(data),
    };
  }

  /**
   * Clear all cached data.
   * Useful for debugging or when cache becomes corrupted.
   * Also resets quota exceeded flag to allow future writes.
   */
  clearAll(): void {
    debugLog("Cache", "Clearing all cached data");

    this.positions.clear();
    this.assignments.clear();
    this.labels.clear();

    // Reset quota flag to allow writes again
    this.quotaExceeded = false;

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

    for (const [k, v] of Object.entries(data.positions)) {
      this.positions.set(k, v);
    }
    for (const [k, v] of Object.entries(data.assignments)) {
      this.assignments.set(k, v);
    }
    for (const [k, v] of Object.entries(data.labels)) {
      this.labels.set(k, v);
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
