import { graphlib, layout as dagreLayout } from "@dagrejs/dagre";
import type { Edge, Node } from "@xyflow/svelte";
import { Position } from "@xyflow/svelte";
import {
  ResourceKind,
  coerceResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createResourceId,
  resourceNameToId,
} from "@rilldata/web-common/features/entity-management/resource-utils";
import type {
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { ResourceNodeData, ResourceMetadata } from "../shared/types";
import { graphCache } from "../shared/cache/position-cache";
import { NODE_CONFIG, DAGRE_CONFIG, EDGE_CONFIG } from "../shared/config";

// Use centralized configuration
const MIN_NODE_WIDTH = NODE_CONFIG.MIN_WIDTH;
const MAX_NODE_WIDTH = NODE_CONFIG.MAX_WIDTH;
const DEFAULT_NODE_HEIGHT = NODE_CONFIG.DEFAULT_HEIGHT;
const AVERAGE_CHAR_WIDTH = NODE_CONFIG.AVERAGE_CHAR_WIDTH;
const CONTENT_PADDING = NODE_CONFIG.CONTENT_PADDING;

// Dagre configuration from centralized config
const DAGRE_NODESEP = DAGRE_CONFIG.NODESEP;
const DAGRE_RANKSEP = DAGRE_CONFIG.RANKSEP;
const DAGRE_EDGESEP = DAGRE_CONFIG.EDGESEP;

// Edge styling from centralized config
const DEFAULT_EDGE_STYLE = EDGE_CONFIG.DEFAULT_STYLE;

// Resource kinds that should be displayed in the graph
const ALLOWED_KINDS = new Set<ResourceKind>([
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
  ResourceKind.Canvas,
]);

function toResourceKind(name?: V1ResourceName): ResourceKind | undefined {
  if (!name?.kind) return undefined;
  return name.kind as ResourceKind;
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

/**
 * Format a V1Schedule into a human-readable description.
 */
function formatScheduleDescription(
  schedule:
    | { cron?: string; tickerSeconds?: number; timeZone?: string }
    | undefined,
): string | undefined {
  if (!schedule) return undefined;
  if (schedule.cron) {
    const tz = schedule.timeZone ? ` (${schedule.timeZone})` : "";
    return `cron: ${schedule.cron}${tz}`;
  }
  if (schedule.tickerSeconds) {
    const seconds = schedule.tickerSeconds;
    if (seconds >= 3600 && seconds % 3600 === 0) {
      const hours = seconds / 3600;
      return `every ${hours}h`;
    }
    if (seconds >= 60 && seconds % 60 === 0) {
      const minutes = seconds / 60;
      return `every ${minutes}m`;
    }
    return `every ${seconds}s`;
  }
  return undefined;
}

/**
 * Infer the actual data source type from a DuckDB path or SQL.
 * DuckDB can read from various cloud storage providers, so we detect the
 * underlying source from the path prefix to show the appropriate icon.
 *
 * The content can be either:
 * - A direct path like "s3://bucket/file.parquet"
 * - A SQL query like "SELECT * FROM read_parquet('s3://bucket/file.parquet')"
 *
 * We search for cloud storage URL patterns anywhere in the content.
 *
 * @param content - The path, URI, or SQL from the DuckDB source/model
 * @returns The inferred connector type, or undefined if no external source detected
 */
function inferDuckDbSourceType(
  content: string | undefined,
): string | undefined {
  if (!content) return undefined;

  const normalized = content.toLowerCase();

  // Check for cloud storage URL patterns anywhere in the content
  // These patterns work for both direct paths and URLs embedded in SQL
  if (normalized.includes("s3://") || normalized.includes("s3a://")) {
    return "s3";
  }
  if (normalized.includes("gs://") || normalized.includes("gcs://")) {
    return "gcs";
  }
  if (
    normalized.includes("azure://") ||
    normalized.includes("az://") ||
    normalized.includes("abfs://") ||
    normalized.includes("abfss://")
  ) {
    return "azure";
  }

  // For HTTP(S), only match if it looks like a data file URL (not documentation links)
  // Check for common data file extensions near the URL
  const httpMatch = normalized.match(/https?:\/\/[^\s'"]+/g);
  if (httpMatch) {
    const dataExtensions = [
      ".parquet",
      ".csv",
      ".json",
      ".ndjson",
      ".jsonl",
      ".xlsx",
      ".xls",
      ".tsv",
    ];
    for (const url of httpMatch) {
      if (dataExtensions.some((ext) => url.includes(ext))) {
        return "https";
      }
    }
  }

  // Check for explicit DuckDB read functions with local paths
  if (
    normalized.includes("read_parquet(") ||
    normalized.includes("read_csv(") ||
    normalized.includes("read_json(") ||
    normalized.includes("read_ndjson(")
  ) {
    return "local_file";
  }

  // No external data source detected (e.g., SQL referencing other models)
  return undefined;
}

/**
 * Extract rich metadata from a resource for badge display.
 */
function extractResourceMetadata(
  resource: V1Resource,
  kind: ResourceKind | undefined,
  allResources: V1Resource[],
): ResourceMetadata {
  const metadata: ResourceMetadata = {};

  // Model/Source metadata
  const model = resource.model;
  const source = resource.source;

  if (model?.spec) {
    const spec = model.spec;
    let connector = spec.inputConnector;

    // For DuckDB connector, infer the actual source type from the path
    if (connector?.toLowerCase() === "duckdb") {
      const inputProps = spec.inputProperties as
        | { path?: string; sql?: string }
        | undefined;
      const path = inputProps?.path || inputProps?.sql;
      connector = inferDuckDbSourceType(path);
    }

    metadata.connector = connector;
    metadata.incremental = spec.incremental;
    metadata.partitioned = Boolean(spec.partitionsResolver);
    metadata.hasSchedule = Boolean(
      spec.refreshSchedule?.cron || spec.refreshSchedule?.tickerSeconds,
    );
    metadata.scheduleDescription = formatScheduleDescription(
      spec.refreshSchedule,
    );
    metadata.retryAttempts = spec.retryAttempts;

    // Check if model is defined via SQL file
    const filePaths = resource.meta?.filePaths ?? [];
    metadata.isSqlModel = filePaths.some((fp) => fp.endsWith(".sql"));
  }

  if (source?.spec) {
    const spec = source.spec;
    let connector = spec.sourceConnector;

    // For DuckDB connector, infer the actual source type from the path
    if (connector?.toLowerCase() === "duckdb") {
      const props = spec.properties as
        | { path?: string; sql?: string }
        | undefined;
      const path = props?.path || props?.sql;
      connector = inferDuckDbSourceType(path);
    }

    metadata.connector = connector;
    metadata.hasSchedule = Boolean(
      spec.refreshSchedule?.cron || spec.refreshSchedule?.tickerSeconds,
    );
    metadata.scheduleDescription = formatScheduleDescription(
      spec.refreshSchedule,
    );
  }

  // Dashboard (Explore/Canvas) theme metadata
  const explore = resource.explore;
  const canvas = resource.canvas;

  if (explore?.spec?.theme && !explore?.spec?.embeddedTheme) {
    metadata.theme = explore.spec.theme;
  }
  if (canvas?.spec?.theme && !canvas?.spec?.embeddedTheme) {
    metadata.theme = canvas.spec.theme;
  }

  // Count alerts and APIs that reference this resource
  const resourceId = createResourceId(resource.meta);
  if (resourceId) {
    let alertCount = 0;
    let apiCount = 0;

    for (const res of allResources) {
      const resKind = res.meta?.name?.kind;
      const refs = res.meta?.refs ?? [];

      // Check if this resource references our target
      const refsTarget = refs.some((ref) => {
        const refId = resourceNameToId(ref);
        return refId === resourceId;
      });

      if (refsTarget) {
        if (resKind === "rill.runtime.v1.Alert") {
          alertCount++;
        } else if (resKind === "rill.runtime.v1.API") {
          apiCount++;
        }
      }
    }

    if (alertCount > 0) metadata.alertCount = alertCount;
    if (apiCount > 0) metadata.apiCount = apiCount;
  }

  return metadata;
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
 * - Filters resources to only allowed kinds (Source, Model, MetricsView, Explore, Canvas)
 *   Note: Sources and Models are merged in the UI (Source is deprecated)
 * - Creates nodes with dynamic widths based on label length
 * - Generates edges based on resource references
 * - Applies Dagre layout with configurable spacing
 * - Caches node positions per namespace for consistent placement
 * - Enforces rank constraints (Explores/Canvas at bottom, Sources/Models flow naturally)
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
      // Sources and Models are merged (Source is deprecated), both treated the same
      case ResourceKind.Source:
      case ResourceKind.Model:
        // No special rank constraint - let them flow naturally in the graph
        rankConstraint = undefined;
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

    const metadata = extractResourceMetadata(resource, kind, resources);

    const nodeDef: Node<ResourceNodeData> = {
      id,
      width: nodeWidth,
      height: DEFAULT_NODE_HEIGHT,
      data: {
        resource,
        kind,
        label,
        metadata,
      },
      type: "resource-node",
      position: { x: 0, y: 0 },
      targetPosition: Position.Top,
      sourcePosition: Position.Bottom,
    };
    nodeDefinitions.set(id, nodeDef);
  }

  // Build adjacency map for edges
  // dependentsMap: sourceId -> Set of dependentIds (outgoing edges from source)
  const dependentsMap = new Map<string, Set<string>>();

  for (const resource of resourceMap.values()) {
    const dependentId = createResourceId(resource.meta);
    if (!dependentId) continue;

    for (const ref of resource.meta?.refs ?? []) {
      const sourceId = resourceNameToId(ref);
      if (!sourceId) continue;
      if (!resourceMap.has(sourceId)) continue;
      if (sourceId === dependentId) continue;

      // Track outgoing edges (source -> dependent)
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
    const cached = opts?.ignoreCache
      ? undefined
      : graphCache.getPosition(posKey);
    node.position = cached ?? computed;
    node.targetPosition = Position.Top;
    node.sourcePosition = Position.Bottom;
    graphCache.setPosition(posKey, node.position);
  }

  // Persist positions for future renders
  graphCache.persist();

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
    graphCache.setLabel(group.id, group.label ?? group.id);
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
    const cachedGroupId = graphCache.getAssignment(id);
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
    const cachedGroupId = graphCache.getAssignment(id);
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
    const label = graphCache.getLabel(syntheticId) ?? "Recovered group";
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
 * Update persistent caches with current grouping state.
 * Stores group labels and resource-to-group assignments.
 */
function updateGroupingCaches(groups: ResourceGraphGrouping[]): void {
  for (const group of groups) {
    if (group.label) graphCache.setLabel(group.id, group.label);
    for (const res of group.resources) {
      const rid = createResourceId(res.meta);
      if (rid) graphCache.setAssignment(rid, group.id);
    }
  }
  graphCache.persist();
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
  filterKind?: ResourceKind | "dashboards",
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
  updateGroupingCaches(groups);

  // If filtering by a specific kind, remove groups that don't contain any resource of that kind
  // Use coerceResourceKind to handle "defined-as-source" models correctly
  // Special case: "dashboards" includes both Explore and Canvas
  if (filterKind) {
    return groups.filter((group) =>
      group.resources.some((r) => {
        const kind = coerceResourceKind(r);
        if (filterKind === "dashboards") {
          return kind === ResourceKind.Explore || kind === ResourceKind.Canvas;
        }
        return kind === filterKind;
      }),
    );
  }

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
    graphCache.setLabel(m.id, m.label);
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
    if (g.label) graphCache.setLabel(g.id, g.label);
    for (const r of g.resources) {
      const rid = createResourceId(r.meta);
      if (rid) graphCache.setAssignment(rid, g.id);
    }
  }

  graphCache.persist();
  return groups;
}
