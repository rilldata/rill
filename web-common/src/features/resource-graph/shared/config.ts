/**
 * Centralized configuration for resource graph layout and visualization.
 *
 * This file consolidates all layout constants, sizing parameters, and visual
 * configuration to make them easier to maintain and adjust.
 */

import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

/**
 * Node sizing configuration
 */
export const NODE_CONFIG = {
  /** Minimum node width in pixels. */
  MIN_WIDTH: 200,
  /** Maximum node width in pixels before text truncates. */
  MAX_WIDTH: 320,
  /** Default height for node: single title row only. */
  DEFAULT_HEIGHT: 36,
  /** Average pixel width per character in node label font. */
  AVERAGE_CHAR_WIDTH: 8.5,
  /** Total horizontal padding within a node (icons, margins, etc.). */
  CONTENT_PADDING: 72,
} as const;

/**
 * Dagre layout spacing configuration.
 *
 * These values were tuned for readability with graphs of 5-50 nodes.
 * Tested with real-world Rill projects containing complex dependency chains.
 */
export const DAGRE_CONFIG = {
  /** Spacing between sibling nodes at the same rank (vertical in LR). */
  NODESEP: 20,
  /** Spacing between graph layers/ranks (horizontal in LR). */
  RANKSEP: 80,
  /** Minimum spacing between edge paths. */
  EDGESEP: 4,
  /** Graph direction: TB (top-to-bottom) or LR (left-to-right). */
  RANKDIR: "LR" as const,
  /** Ranker algorithm: "tight-tree" produces more compact layouts. */
  RANKER: "tight-tree" as const,
  /** Acyclicer algorithm: "greedy" is faster than default. */
  ACYCLICER: "greedy" as const,
} as const;

/**
 * Edge styling and routing configuration
 */
export const EDGE_CONFIG = {
  DEFAULT_STYLE: "stroke:#b1b1b7;stroke-width:1px;opacity:0.85;",
  ERROR_STYLE: "stroke:#ef4444;stroke-width:1.5px;opacity:0.9;",
  HIGHLIGHT_STYLE: "stroke:#3b82f6;stroke-width:2px;opacity:1;",
  DIM_STYLE: "stroke:#b1b1b7;stroke-width:1px;opacity:0.25;",
  DEFAULT_OFFSET: 8,
  MIN_OFFSET: 4,
  MAX_OFFSET: 18,
  BORDER_RADIUS: 6,
  VERTICAL_THRESHOLD_PX: 12,
  OFFSET_SCALING_FACTOR: 10,
} as const;

/**
 * UI layout configuration
 */
export const UI_CONFIG = {
  CARD_HEIGHT_PX: 360,
  DEFAULT_GRID_COLUMNS: 3,
  EXPANDED_HEIGHT_MOBILE: "700px",
  EXPANDED_HEIGHT_DESKTOP: "860px",
} as const;

/**
 * Fit view configuration for graph centering and zoom.
 */
export const FIT_VIEW_CONFIG = {
  PADDING: 0.15,
  MIN_ZOOM: 0.1,
  MAX_ZOOM: 1.25,
  DURATION: 0,
} as const;

/**
 * Performance and optimization configuration.
 */
export const PERFORMANCE_CONFIG = {
  /** Debounce delay for cache writes in milliseconds. */
  CACHE_WRITE_DEBOUNCE_MS: 300,
  /** Seed transition loading indicator duration in milliseconds. */
  SEED_TRANSITION_DELAY_MS: 500,
} as const;

/**
 * Cache size and pruning configuration.
 */
export const CACHE_CONFIG = {
  /** Maximum cache size in bytes (4MB). */
  MAX_SIZE_BYTES: 4 * 1024 * 1024,
  /** Minimum interval between pruning operations in milliseconds. */
  MIN_PRUNE_INTERVAL_MS: 5000,
  /** Percentage of entries to remove during pruning. */
  PRUNE_PERCENTAGE: 0.25,
} as const;

/**
 * Ordering and labels for resource kind sections in the graph UI.
 * Shared by ResourceGraph dropdown and ResourceNodeSelector sidebar.
 */
export const RESOURCE_SECTION_ORDER: ResourceKind[] = [
  ResourceKind.Connector,
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
  ResourceKind.Canvas,
];

export const RESOURCE_SECTION_LABELS: Partial<Record<ResourceKind, string>> = {
  [ResourceKind.Connector]: "OLAP Connector",
  [ResourceKind.Source]: "Source Models",
  [ResourceKind.Model]: "Models",
  [ResourceKind.MetricsView]: "Metric Views",
  [ResourceKind.Explore]: "Explore Dashboards",
  [ResourceKind.Canvas]: "Canvas Dashboards",
};

/**
 * Cache namespace for localStorage.
 */
export const CACHE_NAMESPACE = "rill:resource-graph" as const;

/**
 * Pattern for matching old cache keys during cleanup.
 */
export const CACHE_KEY_PATTERN = /^rill[.:]resource[-.]?[Gg]raph(\.v\d+)?$/;

/**
 * Helper to log debug messages when debug mode is enabled.
 * Enable via: window.__DEBUG_RESOURCE_GRAPH = true
 */
export function debugLog(category: string, message: string, data?: any): void {
  if (
    typeof window === "undefined" ||
    !(window as any).__DEBUG_RESOURCE_GRAPH
  ) {
    return;
  }

  const prefix = `[ResourceGraph:${category}]`;
  if (data !== undefined) {
    console.log(prefix, message, data);
  } else {
    console.log(prefix, message);
  }
}
