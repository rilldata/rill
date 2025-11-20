import { graphlib, layout as dagreLayout } from "@dagrejs/dagre";
import type { Edge, Node } from "@xyflow/svelte";
import { Position } from "@xyflow/svelte";
import {
  ResourceKind,
  coerceResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createResourceId,
  parseResourceId,
  resourceNameToId,
} from "@rilldata/web-common/features/entity-management/resource-utils";
import type {
  V1Resource,
  V1ResourceMeta,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { ResourceNodeData } from "./types";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils/local-storage";

// Node sizing constants
// Chosen to accommodate typical resource names (5-30 characters) without excessive whitespace or text wrapping
const MIN_NODE_WIDTH = 160; // Minimum for short names like "Users" or "Orders"
const MAX_NODE_WIDTH = 320; // Maximum before text wraps; handles names up to ~35 chars
const DEFAULT_NODE_HEIGHT = 56; // Height for single-line text + padding (optimized for 16px font)

// Dagre layout spacing configuration (centralized for cache versioning)
// These values were tuned for readability with graphs of 5-50 nodes.
// Original values (18, 48, 4) were increased by 1.5x to reduce visual density.
// Tested with real-world Rill projects containing complex dependency chains.
const DAGRE_NODESEP = 27; // Horizontal spacing between sibling nodes at the same rank (was 18)
const DAGRE_RANKSEP = 72; // Vertical spacing between graph layers/ranks (was 48)
const DAGRE_EDGESEP = 4; // Minimum spacing between edge paths (rarely matters in practice)

// Default edge styling for non-highlighted edges
const DEFAULT_EDGE_STYLE = "stroke:#b1b1b7;stroke-width:1px;opacity:0.85;";

// Resource kinds that should be displayed in the graph
const ALLOWED_KINDS = new Set<ResourceKind>([
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
]);

// Node width estimation constants (used to dynamically size nodes based on label length)
const AVERAGE_CHAR_WIDTH = 8.5; // Average pixel width per character in the node label font
const CONTENT_PADDING = 72; // Total horizontal padding (icons, margins, etc.) within a node

// Cache last known group assignments so that if a group's anchor
// (e.g., a metrics view) disappears due to an error, we can keep
// its resources together in the expected graph grouping.
const lastGroupAssignments = new Map<string, string>(); // resourceId -> groupId
const lastGroupLabels = new Map<string, string>(); // groupId -> label
const lastPositions = new Map<string, { x: number; y: number }>(); // nodeId -> position
const lastRefs = new Map<string, string[]>(); // dependentId -> [sourceIds]

// Persistent client-side cache (localStorage) to keep positions, grouping, and refs
// Increment CACHE_VERSION when layout algorithm or visual spacing changes significantly
// This forces cache invalidation and prevents stale layouts after code changes.
//
// Version History:
// v1 (initial): Layout with DAGRE_NODESEP=27, RANKSEP=72, basic position caching
//
// Future: Bump version when changing spacing constants, node dimensions, or graph algorithms
const CACHE_VERSION = 1;
const CACHE_NS = `rill.resourceGraph.v${CACHE_VERSION}`;

type PersistedCache = {
  positions: Record<string, { x: number; y: number }>;
  assignments: Record<string, string>;
  labels: Record<string, string>;
  refs: Record<string, string[]>;
};

const DEFAULT_CACHE: PersistedCache = {
  positions: {},
  assignments: {},
  labels: {},
  refs: {},
};

/**
 * Clean up orphaned cache entries from old versions.
 * This prevents localStorage from accumulating stale cache data.
 */
function cleanupOrphanedCaches() {
  try {
    if (typeof window === "undefined" || !window.localStorage) return;

    const pattern = /^rill\.resourceGraph\.v\d+$/;
    const keys = Object.keys(window.localStorage);
    let cleanedCount = 0;

    for (const key of keys) {
      // Match old version keys but not current version
      if (pattern.test(key) && key !== CACHE_NS) {
        window.localStorage.removeItem(key);
        cleanedCount++;
      }
    }

    if (cleanedCount > 0 && typeof console !== "undefined") {
      console.debug(
        `[ResourceGraph] Cleaned up ${cleanedCount} orphaned cache ${cleanedCount === 1 ? "entry" : "entries"}`,
      );
    }
  } catch (error) {
    if (typeof console !== "undefined") {
      console.warn("[ResourceGraph] Failed to cleanup orphaned caches:", error);
    }
  }
}

// Clean up orphaned cache entries on module initialization
cleanupOrphanedCaches();

// Use the project's standard localStorage store pattern
// This handles browser checks, JSON parsing, debouncing, and error handling automatically
const graphCacheStore = localStorageStore<PersistedCache>(
  CACHE_NS,
  DEFAULT_CACHE,
);

// Subscribe to cache changes and sync to in-memory Maps for fast access
graphCacheStore.subscribe((cache) => {
  // Update positions map
  lastPositions.clear();
  for (const [k, v] of Object.entries(cache.positions)) {
    lastPositions.set(k, v);
  }

  // Update assignments map
  lastGroupAssignments.clear();
  for (const [k, v] of Object.entries(cache.assignments)) {
    lastGroupAssignments.set(k, v);
  }

  // Update labels map
  lastGroupLabels.clear();
  for (const [k, v] of Object.entries(cache.labels)) {
    lastGroupLabels.set(k, v);
  }

  // Update refs map
  lastRefs.clear();
  for (const [k, v] of Object.entries(cache.refs)) {
    lastRefs.set(k, v);
  }
});

/**
 * Persist all in-memory caches to localStorage.
 * The store automatically debounces writes (300ms) and handles errors.
 */
function persistAllCaches() {
  graphCacheStore.set({
    positions: Object.fromEntries(lastPositions),
    assignments: Object.fromEntries(lastGroupAssignments),
    labels: Object.fromEntries(lastGroupLabels),
    refs: Object.fromEntries(lastRefs),
  });
}

function toResourceKind(name?: V1ResourceName): ResourceKind | undefined {
  if (!name?.kind) return undefined;
  return name.kind as ResourceKind;
}

/**
 * Create a placeholder resource for a missing or errored dependency.
 * Placeholder resources represent cached nodes that are no longer available
 * but are retained to maintain graph stability and show errors.
 *
 * @param id - Resource ID in format "kind:name"
 * @returns Partial V1Resource with minimal required fields for graph rendering
 */
function makePlaceholderResource(id: string): V1Resource {
  const parsed = parseResourceId(id);
  const name = parsed?.name ?? id;
  const kind = parsed?.kind ?? "model";
  const refIds = lastRefs.get(id) ?? [];
  const refs = refIds
    .map((rid) => parseResourceId(rid))
    .filter((r): r is { kind: string; name: string } => !!r)
    .map((r) => ({ kind: r.kind, name: r.name }));

  // Create a partial resource object with only the fields needed for graph visualization
  // This is a safe type assertion because we're providing all required fields for the graph
  return {
    meta: {
      name: { kind, name },
      refs,
      reconcileError:
        "Resource unavailable due to error or missing spec (cached)",
      hidden: false,
    },
  } as V1Resource;
}

function estimateNodeWidth(label?: string | null) {
  const text = label?.trim() ?? "";
  if (!text.length) return MIN_NODE_WIDTH;
  const segments = text.split(/\s+/).filter(Boolean);
  const longestSegment = segments.length
    ? Math.max(...segments.map((segment) => segment.length))
    : text.length;
  const estimated =
    CONTENT_PADDING +
    Math.max(text.length, longestSegment) * AVERAGE_CHAR_WIDTH;
  return Math.max(
    MIN_NODE_WIDTH,
    Math.min(MAX_NODE_WIDTH, Math.round(estimated)),
  );
}

type BuildGraphOptions = {
  // Namespace for caching node positions. Using a per-graph id avoids reusing
  // coordinates from unrelated graphs, which can place nodes far away.
  positionNs?: string;
  // If true, skip using cached positions and rely on Dagre layout.
  // Helpful to recover from stale caches that can cause overlaps.
  ignoreCache?: boolean;
};

/**
 * Build a directed acyclic graph (DAG) visualization of resource dependencies.
 * Uses Dagre for automatic layout and maintains cached node positions for stability.
 *
 * This function:
 * - Filters resources to only allowed kinds (Source, Model, MetricsView, Explore)
 * - Creates nodes with dynamic widths based on label length
 * - Generates edges based on resource references
 * - Applies Dagre layout with configurable spacing
 * - Caches node positions per namespace for consistent placement
 * - Enforces rank constraints (Sources at top, Explores/Canvas at bottom)
 *
 * @param resources - Array of V1Resource objects to visualize
 * @param opts - Optional configuration for layout and caching
 * @param opts.positionNs - Namespace for position cache (default: "global")
 * @param opts.ignoreCache - If true, recalculates all positions instead of using cache
 * @returns Object containing nodes and edges for SvelteFlow rendering
 *
 * @example
 * const { nodes, edges } = buildResourceGraph(resources, {
 *   positionNs: "dashboard-view",
 *   ignoreCache: false
 * });
 */
export function buildResourceGraph(
  resources: V1Resource[],
  opts?: BuildGraphOptions,
) {
  const positionNs = opts?.positionNs?.trim() || "global";
  const dagreGraph = new graphlib.Graph();
  dagreGraph.setGraph({
    rankdir: "TB",
    // Extreme compactness; overlaps allowed
    nodesep: DAGRE_NODESEP,
    ranksep: DAGRE_RANKSEP,
    edgesep: DAGRE_EDGESEP,
    ranker: "tight-tree",
    acyclicer: "greedy",
  });
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const resourceMap = new Map<string, V1Resource>();
  const nodeDefinitions = new Map<string, Node<ResourceNodeData>>();

  for (const resource of resources) {
    const id = createResourceId(resource.meta);
    if (!id) continue;

    const kind = coerceResourceKind(resource);
    if (!kind || !ALLOWED_KINDS.has(kind)) continue;
    if (resource.meta?.hidden) continue;

    resourceMap.set(id, resource);
    const label = resource.meta?.name?.name ?? "";
    const nodeWidth = estimateNodeWidth(label);
    let rankConstraint: "min" | "max" | undefined;
    switch (kind) {
      case ResourceKind.Source:
        rankConstraint = "min";
        break;
      case ResourceKind.Explore:
      case ResourceKind.Canvas:
        rankConstraint = "max";
        break;
      default:
        rankConstraint = undefined;
    }

    dagreGraph.setNode(id, {
      width: nodeWidth,
      height: DEFAULT_NODE_HEIGHT,
      rank: rankConstraint,
    });

    const nodeDef: Node<ResourceNodeData> = {
      id,
      width: nodeWidth,
      height: DEFAULT_NODE_HEIGHT,
      data: {
        resource,
        kind,
        label,
      },
      type: "resource-node",
      position: { x: 0, y: 0 },
      targetPosition: Position.Top,
      sourcePosition: Position.Bottom,
    };
    nodeDefinitions.set(id, nodeDef);
  }

  const dependentsMap = new Map<string, Set<string>>();

  for (const resource of resourceMap.values()) {
    const dependentId = createResourceId(resource.meta);
    if (!dependentId) continue;

    for (const ref of resource.meta?.refs ?? []) {
      const sourceId = resourceNameToId(ref);
      if (!sourceId) continue;
      if (!resourceMap.has(sourceId)) continue;
      if (sourceId === dependentId) continue;

      if (!dependentsMap.has(sourceId)) dependentsMap.set(sourceId, new Set());
      dependentsMap.get(sourceId)!.add(dependentId);
    }
  }

  const edgeIds = new Set<string>();
  const edges: Edge[] = [];

  for (const [sourceId, dependents] of dependentsMap) {
    if (!dependents?.size) continue;
    for (const dependentId of dependents) {
      if (!resourceMap.has(sourceId) || !resourceMap.has(dependentId)) continue;
      const edgeId = `${sourceId}->${dependentId}`;
      if (edgeIds.has(edgeId)) continue;
      edgeIds.add(edgeId);
      dagreGraph.setEdge(sourceId, dependentId);
      const edge: Edge = {
        id: edgeId,
        source: sourceId,
        target: dependentId,
        animated: false,
        type: "smoothstep",
        style: DEFAULT_EDGE_STYLE,
      } as Edge;
      edges.push(edge);
    }
  }

  const nodes: Node<ResourceNodeData>[] = Array.from(nodeDefinitions.values());

  dagreLayout(dagreGraph);

  for (const node of nodes) {
    const dagreNode = dagreGraph.node(node.id);
    if (!dagreNode) continue;
    const nodeWidth = node.width ?? MIN_NODE_WIDTH;
    const computed = {
      x: dagreNode.x - nodeWidth / 2,
      y: dagreNode.y - (node.height ?? DEFAULT_NODE_HEIGHT) / 2,
    };
    const posKey = `${positionNs}|${node.id}`;
    const cached = opts?.ignoreCache ? undefined : lastPositions.get(posKey);
    node.position = cached ?? computed;
    node.targetPosition = Position.Top;
    node.sourcePosition = Position.Bottom;
    lastPositions.set(posKey, node.position);
  }

  // Persist positions for future renders
  persistAllCaches();

  return { nodes, edges };
}

export type ResourceGraphGrouping = {
  id: string;
  resources: V1Resource[];
  label?: string;
};

// Build directed adjacency: incoming (sources) and outgoing (dependents)
function buildDirectedAdjacency(resources: Map<string, V1Resource>) {
  const incoming = new Map<string, Set<string>>(); // node <- sources
  const outgoing = new Map<string, Set<string>>(); // node -> dependents
  for (const id of resources.keys()) {
    if (!incoming.has(id)) incoming.set(id, new Set());
    if (!outgoing.has(id)) outgoing.set(id, new Set());
  }
  for (const resource of resources.values()) {
    const dependentId = createResourceId(resource.meta);
    if (!dependentId) continue;
    for (const ref of resource.meta?.refs ?? []) {
      const sourceId = createResourceId({ name: ref });
      if (!sourceId) continue;
      // Record refs for persistence even if source is not currently present
      const existing = lastRefs.get(dependentId) ?? [];
      if (!existing.includes(sourceId))
        lastRefs.set(dependentId, [...existing, sourceId]);
      if (!resources.has(sourceId)) continue;
      if (!incoming.has(dependentId)) incoming.set(dependentId, new Set());
      if (!outgoing.has(sourceId)) outgoing.set(sourceId, new Set());
      incoming.get(dependentId)!.add(sourceId);
      outgoing.get(sourceId)!.add(dependentId);
    }
  }
  return { incoming, outgoing };
}

/**
 * Graph Traversal Strategy
 *
 * We use breadth-first search (BFS) to find all nodes reachable from seed nodes.
 * This includes both upstream dependencies (sources) and downstream dependents.
 *
 * Performance Optimization:
 * - Uses index-based queue iteration instead of array.shift() to avoid O(n²) performance
 * - array.shift() is O(n) per call, making traditional BFS O(n²) for large graphs
 * - Index-based approach is O(1) per iteration, achieving true O(V + E) complexity
 *
 * Complexity: O(V + E) where V = nodes, E = edges
 * Space: O(V) for visited set and queue
 *
 * The visited set prevents infinite loops in case of circular dependencies
 * (which shouldn't exist in a true DAG but we handle defensively).
 */

// Traverse only upstream via incoming edges
function traverseUpstream(seedId: string, incoming: Map<string, Set<string>>) {
  const visited = new Set<string>();
  const queue = [seedId];
  let queueIndex = 0;
  while (queueIndex < queue.length) {
    const current = queue[queueIndex++];
    if (visited.has(current)) continue;
    visited.add(current);
    const parents = incoming.get(current);
    if (!parents?.size) continue;
    for (const p of parents) if (!visited.has(p)) queue.push(p);
  }
  return visited;
}

// Traverse only downstream via outgoing edges
function traverseDownstream(
  seedId: string,
  outgoing: Map<string, Set<string>>,
) {
  const visited = new Set<string>();
  const queue = [seedId];
  let queueIndex = 0;
  while (queueIndex < queue.length) {
    const current = queue[queueIndex++];
    if (visited.has(current)) continue;
    visited.add(current);
    const children = outgoing.get(current);
    if (!children?.size) continue;
    for (const c of children) if (!visited.has(c)) queue.push(c);
  }
  return visited;
}

/**
 * Build a map of visible resources that are allowed in the graph.
 * Filters out hidden resources and disallowed kinds.
 */
function buildVisibleResourceMap(
  resources: V1Resource[],
): Map<string, V1Resource> {
  const resourceMap = new Map<string, V1Resource>();
  for (const resource of resources) {
    const id = createResourceId(resource.meta);
    if (!id) continue;
    const kind = toResourceKind(resource.meta?.name);
    if (!kind || !ALLOWED_KINDS.has(kind)) continue;
    if (resource.meta?.hidden) continue;
    resourceMap.set(id, resource);
  }
  return resourceMap;
}

/**
 * Normalize and deduplicate seed identifiers.
 * Converts V1ResourceName objects to string IDs and removes duplicates.
 */
function normalizeSeeds(seeds: (string | V1ResourceName)[]): string[] {
  const toSeedId = (seed: string | V1ResourceName) =>
    typeof seed === "string" ? seed : createResourceId({ name: seed });

  return seeds
    .map((s) => toSeedId(s))
    .filter((id): id is string => !!id)
    .filter((id, idx, arr) => arr.indexOf(id) === idx);
}

/**
 * Create initial groups by traversing upstream and downstream from each seed.
 * Returns groups, a lookup map, and the set of assigned resource IDs.
 */
function createSeedBasedGroups(
  normalizedSeeds: string[],
  resourceMap: Map<string, V1Resource>,
  incoming: Map<string, Set<string>>,
  outgoing: Map<string, Set<string>>,
): {
  groups: ResourceGraphGrouping[];
  groupById: Map<string, ResourceGraphGrouping>;
  assigned: Set<string>;
} {
  const groups: ResourceGraphGrouping[] = [];
  const groupById = new Map<string, ResourceGraphGrouping>();
  const assigned = new Set<string>();

  for (const seedId of normalizedSeeds) {
    // Directed closure: only upstream via incoming and only downstream via outgoing
    const upIds = traverseUpstream(seedId, incoming);
    const downIds = traverseDownstream(seedId, outgoing);
    const componentIds = new Set<string>([...upIds, ...downIds]);

    const componentResources = Array.from(componentIds)
      .map((resourceId) => resourceMap.get(resourceId))
      .filter((res): res is V1Resource => !!res);
    if (!componentResources.length) continue;

    const label = resourceMap.get(seedId)?.meta?.name?.name ?? seedId;
    const group: ResourceGraphGrouping = {
      id: seedId,
      resources: componentResources,
      label,
    };
    groups.push(group);
    groupById.set(group.id, group);
    lastGroupLabels.set(group.id, group.label ?? group.id);
    for (const resId of componentIds) assigned.add(resId);
  }

  return { groups, groupById, assigned };
}

/**
 * Attempt to assign unassigned resources to their previously cached group.
 * This maintains grouping stability when resources move between groups.
 */
function assignUnassignedResourcesToCachedGroups(
  resourceMap: Map<string, V1Resource>,
  groupById: Map<string, ResourceGraphGrouping>,
  assigned: Set<string>,
): void {
  const unassignedIds = Array.from(resourceMap.keys()).filter(
    (id) => !assigned.has(id),
  );
  for (const id of unassignedIds) {
    const cachedGroupId = lastGroupAssignments.get(id);
    if (!cachedGroupId) continue;
    const existingGroup = groupById.get(cachedGroupId);
    if (!existingGroup) continue;
    const res = resourceMap.get(id);
    if (!res) continue;
    existingGroup.resources.push(res);
    assigned.add(id);
  }
}

/**
 * Create synthetic groups for resources whose anchor (seed) disappeared due to errors.
 * Uses cached group labels to maintain continuity.
 */
function createSyntheticGroupsForMissingAnchors(
  resourceMap: Map<string, V1Resource>,
  groupById: Map<string, ResourceGraphGrouping>,
  groups: ResourceGraphGrouping[],
  assigned: Set<string>,
): void {
  const stillUnassignedIds = Array.from(resourceMap.keys()).filter(
    (id) => !assigned.has(id),
  );
  const syntheticGroups = new Map<string, V1Resource[]>();

  // Collect resources that belong to missing groups
  for (const id of stillUnassignedIds) {
    const cachedGroupId = lastGroupAssignments.get(id);
    if (!cachedGroupId) continue;
    if (groupById.has(cachedGroupId)) continue;
    if (!syntheticGroups.has(cachedGroupId))
      syntheticGroups.set(cachedGroupId, []);
    const res = resourceMap.get(id);
    if (res) syntheticGroups.get(cachedGroupId)!.push(res);
  }

  // Create groups for collected resources
  for (const [syntheticId, syntheticResources] of syntheticGroups) {
    if (!syntheticResources.length) continue;
    const label = lastGroupLabels.get(syntheticId) ?? "Recovered group";
    const group: ResourceGraphGrouping = {
      id: syntheticId,
      resources: syntheticResources,
      label,
    };
    groups.push(group);
    groupById.set(group.id, group);
    for (const res of syntheticResources) {
      const rid = createResourceId(res.meta);
      if (rid) assigned.add(rid);
    }
  }
}

/**
 * Add placeholder resources for cached resources that are temporarily missing.
 * This preserves graph structure when resources have errors.
 */
function addPlaceholdersForMissingResources(
  resourceMap: Map<string, V1Resource>,
  groupById: Map<string, ResourceGraphGrouping>,
  assigned: Set<string>,
): void {
  const presentIds = new Set(resourceMap.keys());
  for (const [rid, gid] of lastGroupAssignments) {
    const grp = groupById.get(gid);
    if (!grp) continue;
    if (presentIds.has(rid)) continue;
    if (grp.resources.some((r) => createResourceId(r.meta) === rid)) continue;
    grp.resources.push(makePlaceholderResource(rid));
    assigned.add(rid);
  }
}

/**
 * Update persistent caches with current grouping state.
 * Stores group labels and resource-to-group assignments.
 */
function updateGroupingCaches(groups: ResourceGraphGrouping[]): void {
  for (const group of groups) {
    if (group.label) lastGroupLabels.set(group.id, group.label);
    for (const res of group.resources) {
      const rid = createResourceId(res.meta);
      if (rid) lastGroupAssignments.set(rid, group.id);
    }
  }
  persistAllCaches();
}

/**
 * Partition resources into groups based on seed resources.
 * Each seed generates a group containing the seed and all resources
 * connected to it (both upstream sources and downstream dependents).
 *
 * This function implements a directed graph traversal strategy:
 * - For each seed, traverses upstream (sources/dependencies) and downstream (dependents)
 * - Creates separate groups for each seed's connected subgraph
 * - Maintains grouping stability across renders using localStorage cache
 * - Handles missing resources by creating placeholder nodes
 * - Preserves groups even when anchor resources temporarily disappear due to errors
 *
 * Use cases:
 * - Viewing dependencies of specific metrics or models
 * - Drilling down into resource lineage
 * - Isolating subgraphs for focused analysis
 *
 * @param resources - All resources to consider for grouping
 * @param seeds - Seed identifiers (can be strings like "model:orders" or V1ResourceName objects)
 * @returns Array of resource groups, one per seed (plus recovered groups from cache)
 *
 * @example
 * // View graphs for specific metrics
 * const groups = partitionResourcesBySeeds(allResources, [
 *   "rill.runtime.v1.MetricsView:revenue",
 *   "rill.runtime.v1.Model:users"
 * ]);
 *
 * @example
 * // Using V1ResourceName format
 * const groups = partitionResourcesBySeeds(allResources, [
 *   { kind: ResourceKind.MetricsView, name: "revenue" }
 * ]);
 */
export function partitionResourcesBySeeds(
  resources: V1Resource[],
  seeds: (string | V1ResourceName)[],
): ResourceGraphGrouping[] {
  const resourceMap = buildVisibleResourceMap(resources);
  const { incoming, outgoing } = buildDirectedAdjacency(resourceMap);
  const normalizedSeeds = normalizeSeeds(seeds);

  const { groups, groupById, assigned } = createSeedBasedGroups(
    normalizedSeeds,
    resourceMap,
    incoming,
    outgoing,
  );

  assignUnassignedResourcesToCachedGroups(resourceMap, groupById, assigned);
  createSyntheticGroupsForMissingAnchors(
    resourceMap,
    groupById,
    groups,
    assigned,
  );
  addPlaceholdersForMissingResources(resourceMap, groupById, assigned);
  updateGroupingCaches(groups);

  return groups;
}

/**
 * Partition resources into groups based on MetricsView resources.
 * Creates one group per MetricsView containing all connected resources.
 *
 * This function implements an undirected graph traversal strategy:
 * - Identifies all MetricsView resources as group anchors
 * - For each MetricsView, traverses the entire connected component (undirected)
 * - Creates groups sorted alphabetically by MetricsView name
 * - Collects orphaned resources (not connected to any MetricsView) into "Other resources" groups
 * - Maintains grouping stability using localStorage cache
 *
 * Use cases:
 * - Default view showing all metrics and their dependencies
 * - Overview of project structure organized by business metrics
 * - Discovering isolated resource components
 *
 * @param resources - All resources to partition into groups
 * @returns Array of resource groups, one per MetricsView plus any orphaned components
 *
 * @example
 * // Create default metric-based grouping
 * const groups = partitionResourcesByMetrics(allResources);
 * // Result: [
 * //   { id: "...:revenue", label: "revenue", resources: [...] },
 * //   { id: "...:user_stats", label: "user_stats", resources: [...] },
 * //   { id: "...:some_model", label: "Other resources", resources: [...] }
 * // ]
 */
export function partitionResourcesByMetrics(
  resources: V1Resource[],
): ResourceGraphGrouping[] {
  // NETWORK approach: build undirected adjacency and group by connected
  // components rooted at metrics views.
  const resourceMap = new Map<string, V1Resource>();
  const adjacency = new Map<string, Set<string>>();

  for (const res of resources) {
    const id = createResourceId(res.meta);
    if (!id) continue;
    const kind = toResourceKind(res.meta?.name);
    if (!kind || !ALLOWED_KINDS.has(kind)) continue;
    if (res.meta?.hidden) continue;
    resourceMap.set(id, res);
    if (!adjacency.has(id)) adjacency.set(id, new Set());
  }

  for (const res of resourceMap.values()) {
    const dependentId = createResourceId(res.meta);
    if (!dependentId) continue;
    for (const ref of res.meta?.refs ?? []) {
      const sourceId = createResourceId({ name: ref });
      if (!sourceId) continue;
      const existing = lastRefs.get(dependentId) ?? [];
      if (!existing.includes(sourceId))
        lastRefs.set(dependentId, [...existing, sourceId]);
      if (!resourceMap.has(sourceId)) continue;
      if (!adjacency.has(sourceId)) adjacency.set(sourceId, new Set());
      if (!adjacency.has(dependentId)) adjacency.set(dependentId, new Set());
      adjacency.get(dependentId)!.add(sourceId);
      adjacency.get(sourceId)!.add(dependentId);
    }
  }

  const metricSeeds = Array.from(resourceMap.entries())
    .filter(
      ([, r]) => toResourceKind(r.meta?.name) === ResourceKind.MetricsView,
    )
    .map(([id, r]) => ({ id, label: r.meta?.name?.name ?? id }))
    .sort((a, b) => a.label.localeCompare(b.label));

  // Helper for undirected traversal
  const traverseConnected = (startId: string) => {
    const visited = new Set<string>();
    const queue = [startId];
    let queueIndex = 0;
    while (queueIndex < queue.length) {
      const cur = queue[queueIndex++];
      if (visited.has(cur)) continue;
      visited.add(cur);
      const nbrs = adjacency.get(cur);
      if (!nbrs?.size) continue;
      for (const nb of nbrs) if (!visited.has(nb)) queue.push(nb);
    }
    return visited;
  };

  const groups: ResourceGraphGrouping[] = [];
  const assigned = new Set<string>();

  for (const m of metricSeeds) {
    const ids = traverseConnected(m.id);
    if (!ids.size) continue;
    const resourcesInGroup = Array.from(ids)
      .map((rid) => resourceMap.get(rid))
      .filter((x): x is V1Resource => !!x);
    if (!resourcesInGroup.length) continue;
    groups.push({ id: m.id, resources: resourcesInGroup, label: m.label });
    for (const rid of ids) assigned.add(rid);
    lastGroupLabels.set(m.id, m.label);
  }

  // If there are resources not connected to any metrics view, group remaining components.
  const remaining = Array.from(resourceMap.keys()).filter(
    (id) => !assigned.has(id),
  );
  const remainingSet = new Set(remaining);
  while (remainingSet.size) {
    const seed = remainingSet.values().next().value as string;
    const ids = traverseConnected(seed);
    for (const rid of ids) remainingSet.delete(rid);
    const resourcesInGroup = Array.from(ids)
      .map((rid) => resourceMap.get(rid))
      .filter((x): x is V1Resource => !!x);
    if (resourcesInGroup.length) {
      groups.push({
        id: seed,
        resources: resourcesInGroup,
        label: "Other resources",
      });
    }
  }

  // Persist grouping assignments
  for (const g of groups) {
    if (g.label) lastGroupLabels.set(g.id, g.label);
    for (const r of g.resources) {
      const rid = createResourceId(r.meta);
      if (rid) lastGroupAssignments.set(rid, g.id);
    }
  }

  persistAllCaches();
  return groups;
}
