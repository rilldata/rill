/**
 * Centralized configuration for resource graph layout and visualization.
 *
 * This file consolidates all layout constants, sizing parameters, and visual
 * configuration to make them easier to maintain and adjust.
 */

/**
 * Node sizing configuration
 */
export const NODE_CONFIG = {
  /**
   * Minimum node width in pixels.
   * Chosen to accommodate short names like "Users" or "Orders".
   */
  MIN_WIDTH: 200,

  /**
   * Maximum node width in pixels before text wraps.
   * Handles names up to ~35 characters.
   */
  MAX_WIDTH: 320,

  /**
   * Default height for node: title row + 2 content rows.
   */
  DEFAULT_HEIGHT: 76,

  /**
   * Average pixel width per character in node label font.
   * Used for dynamic width estimation.
   */
  AVERAGE_CHAR_WIDTH: 8.5,

  /**
   * Total horizontal padding within a node (icons, margins, etc.).
   */
  CONTENT_PADDING: 72,
} as const;

/**
 * Dagre layout spacing configuration.
 *
 * These values were tuned for readability with graphs of 5-50 nodes.
 * Original values (18, 48, 4) were increased by 1.5x to reduce visual density.
 * Tested with real-world Rill projects containing complex dependency chains.
 */
export const DAGRE_CONFIG = {
  /**
   * Horizontal spacing between sibling nodes at the same rank.
   */
  NODESEP: 27,

  /**
   * Vertical spacing between graph layers/ranks.
   * Sized to accommodate taller nodes with metadata rows.
   */
  RANKSEP: 96,

  /**
   * Minimum spacing between edge paths.
   * Rarely matters in practice but prevents edge overlap.
   */
  EDGESEP: 4,

  /**
   * Graph direction: TB (top-to-bottom) or LR (left-to-right).
   */
  RANKDIR: "TB" as const,

  /**
   * Ranker algorithm: "tight-tree" | "longest-path" | "network-simplex".
   * "tight-tree" produces more compact layouts.
   */
  RANKER: "tight-tree" as const,

  /**
   * Acyclicer algorithm: "greedy" is faster than default.
   */
  ACYCLICER: "greedy" as const,
} as const;

/**
 * Edge styling and routing configuration
 */
export const EDGE_CONFIG = {
  /**
   * Default edge style for non-highlighted edges.
   */
  DEFAULT_STYLE: "stroke:#b1b1b7;stroke-width:1px;opacity:0.85;",

  /**
   * Style for highlighted/selected edge paths.
   */
  HIGHLIGHT_STYLE: "stroke:#3b82f6;stroke-width:2px;opacity:1;",

  /**
   * Style for dimmed edges when selection exists.
   */
  DIM_STYLE: "stroke:#b1b1b7;stroke-width:1px;opacity:0.25;",

  /**
   * Default offset for edge routing (how far edges curve).
   */
  DEFAULT_OFFSET: 8,

  /**
   * Minimum offset for nearly-vertical edges.
   */
  MIN_OFFSET: 4,

  /**
   * Maximum offset for widely-spaced nodes.
   */
  MAX_OFFSET: 18,

  /**
   * Border radius for edge corner rounding.
   */
  BORDER_RADIUS: 6,

  /**
   * Treat edge as vertical if horizontal distance is less than this.
   */
  VERTICAL_THRESHOLD_PX: 12,

  /**
   * Divides vertical distance to compute dynamic offset.
   */
  OFFSET_SCALING_FACTOR: 10,
} as const;

/**
 * UI layout configuration
 */
export const UI_CONFIG = {
  /**
   * Default card height in pixels.
   * Sized to fit 3x3 grid comfortably on standard displays.
   */
  CARD_HEIGHT_PX: 360,

  /**
   * Default grid columns for graph cards.
   */
  DEFAULT_GRID_COLUMNS: 3,

  /**
   * Expanded graph height on mobile devices.
   */
  EXPANDED_HEIGHT_MOBILE: "700px",

  /**
   * Expanded graph height on desktop devices.
   */
  EXPANDED_HEIGHT_DESKTOP: "860px",
} as const;

/**
 * Fit view configuration for graph centering and zoom.
 */
export const FIT_VIEW_CONFIG = {
  /**
   * Padding around graph as percentage (0.15 = 15%).
   */
  PADDING: 0.15,

  /**
   * Minimum zoom level.
   */
  MIN_ZOOM: 0.1,

  /**
   * Maximum zoom level.
   */
  MAX_ZOOM: 1.25,

  /**
   * Animation duration in milliseconds.
   * Set to 0 to disable animation on initial render (removes distracting pan/zoom effect).
   */
  DURATION: 0,
} as const;

/**
 * Performance and optimization configuration.
 */
export const PERFORMANCE_CONFIG = {
  /**
   * Debounce delay for cache writes in milliseconds.
   * Prevents excessive localStorage writes.
   */
  CACHE_WRITE_DEBOUNCE_MS: 300,

  /**
   * Maximum number of groups to render by default.
   * null means no limit.
   */
  DEFAULT_MAX_GROUPS: null as number | null,

  /**
   * Delay before showing loading spinner (prevents flash).
   */
  LOADING_SPINNER_DELAY_MS: 200,

  /**
   * Seed transition loading indicator duration.
   */
  SEED_TRANSITION_DELAY_MS: 150,
} as const;

/**
 * Cache size and pruning configuration.
 */
export const CACHE_CONFIG = {
  /**
   * Maximum cache size in bytes (4MB).
   * localStorage typically has 5-10MB quota, leaving room for other data.
   */
  MAX_SIZE_BYTES: 4 * 1024 * 1024,

  /**
   * Minimum interval between pruning operations in milliseconds.
   * Prevents thrashing when cache is near limit.
   */
  MIN_PRUNE_INTERVAL_MS: 5000,

  /**
   * Percentage of entries to remove during pruning (0.25 = 25%).
   */
  PRUNE_PERCENTAGE: 0.25,
} as const;

/**
 * Debug mode configuration.
 * Set window.__DEBUG_RESOURCE_GRAPH = true to enable.
 */
export const DEBUG_CONFIG = {
  /**
   * Check if debug mode is enabled.
   */
  get enabled(): boolean {
    return (
      typeof window !== "undefined" &&
      (window as any).__DEBUG_RESOURCE_GRAPH === true
    );
  },

  /**
   * Log graph operations to console.
   */
  logOperations: true,

  /**
   * Log cache operations.
   */
  logCache: true,

  /**
   * Log performance metrics.
   */
  logPerformance: true,

  /**
   * Expose internal state on window for debugging.
   */
  exposeInternals: true,
} as const;

/**
 * Cache namespace for localStorage.
 *
 * Uses `rill:` prefix for namespacing since Rill Developer runs on localhost:9009,
 * which may be shared with other applications. This follows the pattern established
 * by `rill:theme` for consistency.
 *
 * Cache invalidation is handled via the debug utility `window.__RESOURCE_GRAPH_CACHE.clearAll()`
 * rather than versioning, keeping the key simple and user-friendly.
 */
export const CACHE_NAMESPACE = "rill:resource-graph" as const;

/**
 * Pattern for matching old cache keys during cleanup.
 * Matches both old versioned keys (rill.resourceGraph.v1, rill.resourceGraph.v2)
 * and ensures cleanup during migration to new key format.
 */
export const CACHE_KEY_PATTERN = /^rill[.:]resource[-.]?[Gg]raph(\.v\d+)?$/;

/**
 * Helper to log debug messages only when debug mode is enabled.
 */
export function debugLog(category: string, message: string, data?: any): void {
  if (!DEBUG_CONFIG.enabled) return;

  const prefix = `[ResourceGraph:${category}]`;
  if (data !== undefined) {
    console.log(prefix, message, data);
  } else {
    console.log(prefix, message);
  }
}

/**
 * Helper to measure and log performance metrics.
 */
export function debugPerf<T>(operation: string, fn: () => T): T {
  if (!DEBUG_CONFIG.enabled || !DEBUG_CONFIG.logPerformance) {
    return fn();
  }

  const start = performance.now();
  const result = fn();
  const duration = performance.now() - start;

  console.log(
    `[ResourceGraph:Perf] ${operation} took ${duration.toFixed(2)}ms`,
  );

  return result;
}
