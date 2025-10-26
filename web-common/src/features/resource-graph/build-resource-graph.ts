import { graphlib, layout as dagreLayout } from "@dagrejs/dagre";
import type { Edge, Node } from "@xyflow/svelte";
import { Position } from "@xyflow/svelte";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type {
  V1Resource,
  V1ResourceMeta,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { ResourceNodeData } from "./types";

const MIN_NODE_WIDTH = 180;
const MAX_NODE_WIDTH = 360;
const DEFAULT_NODE_HEIGHT = 72;
// Softer, less intrusive edges
const DEFAULT_EDGE_STYLE = "stroke:#b1b1b7;stroke-width:1px;opacity:0.85;";
const ALLOWED_KINDS = new Set<ResourceKind>([
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
]);
const AVERAGE_CHAR_WIDTH = 8.5;
const CONTENT_PADDING = 96;

// Cache last known group assignments so that if a group's anchor
// (e.g., a metrics view) disappears due to an error, we can keep
// its resources together in the expected graph grouping.
const lastGroupAssignments = new Map<string, string>(); // resourceId -> groupId
const lastGroupLabels = new Map<string, string>(); // groupId -> label

// Persistent client-side cache (localStorage) to keep positions, grouping, and refs
const CACHE_NS = "rill.resourceGraph.v1";
type PositionsCache = Record<string, { x: number; y: number }>;
type AssignmentsCache = Record<string, string>;
type LabelsCache = Record<string, string>;
type RefsCache = Record<string, string[]>; // dependentId -> [sourceIds]
type PersistedCache = {
  positions?: PositionsCache;
  assignments?: AssignmentsCache;
  labels?: LabelsCache;
  refs?: RefsCache;
};

function hasLocalStorage() {
  try {
    return typeof window !== "undefined" && !!window.localStorage;
  } catch {
    return false;
  }
}

function loadPersistedCache(): PersistedCache {
  if (!hasLocalStorage()) return {};
  try {
    const raw = window.localStorage.getItem(CACHE_NS);
    if (!raw) return {};
    const data = JSON.parse(raw);
    return typeof data === "object" && data ? data : {};
  } catch {
    return {};
  }
}

function savePersistedCache(cache: PersistedCache) {
  if (!hasLocalStorage()) return;
  try {
    window.localStorage.setItem(CACHE_NS, JSON.stringify(cache));
  } catch {
    // ignore
  }
}

const persisted = loadPersistedCache();
const lastPositions: Map<string, { x: number; y: number }> = new Map(
  Object.entries(persisted.positions ?? {}),
);
for (const [rid, gid] of Object.entries(persisted.assignments ?? {})) {
  lastGroupAssignments.set(rid, gid);
}
for (const [gid, label] of Object.entries(persisted.labels ?? {})) {
  lastGroupLabels.set(gid, label);
}
const lastRefs: Map<string, string[]> = new Map(
  Object.entries(persisted.refs ?? {}),
);

function persistAllCaches() {
  const positions: PositionsCache = {};
  for (const [k, v] of lastPositions) positions[k] = v;
  const assignments: AssignmentsCache = {};
  for (const [k, v] of lastGroupAssignments) assignments[k] = v;
  const labels: LabelsCache = {};
  for (const [k, v] of lastGroupLabels) labels[k] = v;
  const refs: RefsCache = {};
  for (const [k, v] of lastRefs) refs[k] = v;
  savePersistedCache({ positions, assignments, labels, refs });
}

function toNodeId(meta?: V1ResourceMeta): string | undefined {
  if (!meta?.name?.name || !meta?.name?.kind) return undefined;
  return `${meta.name.kind}:${meta.name.name}`;
}

function toResourceKind(name?: V1ResourceName): ResourceKind | undefined {
  if (!name?.kind) return undefined;
  return name.kind as ResourceKind;
}

function parseNodeId(id: string): { kind: string; name: string } | null {
  const idx = id.indexOf(":");
  if (idx <= 0) return null;
  const kind = id.slice(0, idx);
  const name = id.slice(idx + 1);
  if (!kind || !name) return null;
  return { kind, name };
}

function makePlaceholderResource(id: string, label?: string): V1Resource {
  const parsed = parseNodeId(id);
  const name = parsed?.name ?? id;
  const kind = parsed?.kind ?? "model";
  const refIds = lastRefs.get(id) ?? [];
  const refs = refIds
    .map((rid) => parseNodeId(rid))
    .filter((r): r is { kind: string; name: string } => !!r)
    .map((r) => ({ kind: r.kind, name: r.name }));
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
    CONTENT_PADDING + Math.max(text.length, longestSegment) * AVERAGE_CHAR_WIDTH;
  return Math.max(
    MIN_NODE_WIDTH,
    Math.min(MAX_NODE_WIDTH, Math.round(estimated)),
  );
}

export function buildResourceGraph(resources: V1Resource[]) {
  const dagreGraph = new graphlib.Graph();
  dagreGraph.setGraph({
    rankdir: "TB",
    nodesep: 320,
    ranksep: 240,
    edgesep: 80,
    acyclicer: "greedy",
  });
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const resourceMap = new Map<string, V1Resource>();
  const nodeDefinitions = new Map<string, Node<ResourceNodeData>>();

  for (const resource of resources) {
    const id = toNodeId(resource.meta);
    if (!id) continue;

    const kind = toResourceKind(resource.meta?.name);
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
    const dependentId = toNodeId(resource.meta);
    if (!dependentId) continue;

    for (const ref of resource.meta?.refs ?? []) {
      const sourceId = toNodeId({ name: ref });
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
    const cached = lastPositions.get(node.id);
    node.position = cached ?? computed;
    node.targetPosition = Position.Top;
    node.sourcePosition = Position.Bottom;
    lastPositions.set(node.id, node.position);
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

export function partitionResourcesByMetrics(
  resources: V1Resource[],
): ResourceGraphGrouping[] {
  const resourceMap = new Map<string, V1Resource>();
  const adjacency = new Map<string, Set<string>>();

  for (const resource of resources) {
    const id = toNodeId(resource.meta);
    if (!id) continue;

    const kind = toResourceKind(resource.meta?.name);
    if (!kind || !ALLOWED_KINDS.has(kind)) continue;
    if (resource.meta?.hidden) continue;

    resourceMap.set(id, resource);
    if (!adjacency.has(id)) adjacency.set(id, new Set());
  }

  for (const resource of resourceMap.values()) {
    const dependentId = toNodeId(resource.meta);
    if (!dependentId) continue;

    for (const ref of resource.meta?.refs ?? []) {
      const sourceId = toNodeId({ name: ref });
      if (!sourceId) continue;
      // Record refs for persistence even if source is not currently present
      const existing = lastRefs.get(dependentId) ?? [];
      if (!existing.includes(sourceId)) lastRefs.set(dependentId, [...existing, sourceId]);
      if (!resourceMap.has(sourceId)) continue;

      adjacency.get(dependentId)!.add(sourceId);
      if (!adjacency.has(sourceId)) adjacency.set(sourceId, new Set());
      adjacency.get(sourceId)!.add(dependentId);
    }
  }

  const metricIds = Array.from(resourceMap.entries())
    .filter(([, resource]) => toResourceKind(resource.meta?.name) === ResourceKind.MetricsView)
    .map(([id, resource]) => ({
      id,
      label: resource.meta?.name?.name ?? id,
    }))
    .sort((a, b) => a.label.localeCompare(b.label));

  const groups: ResourceGraphGrouping[] = [];
  const groupById = new Map<string, ResourceGraphGrouping>();
  const assigned = new Set<string>();

  const traverseConnected = (startId: string) => {
    const visited = new Set<string>();
    const queue = [startId];

    while (queue.length) {
      const current = queue.shift()!;
      if (visited.has(current)) continue;
      visited.add(current);

      const neighbors = adjacency.get(current);
      if (!neighbors?.size) continue;
      for (const neighborId of neighbors) {
        if (!visited.has(neighborId)) queue.push(neighborId);
      }
    }

    return visited;
  };

  for (const metric of metricIds) {
    const componentIds = traverseConnected(metric.id);

    const componentResources = Array.from(componentIds)
      .map((resourceId) => resourceMap.get(resourceId))
      .filter((res): res is V1Resource => !!res);

    if (componentResources.length) {
      const group: ResourceGraphGrouping = {
        id: metric.id,
        resources: componentResources,
        label: metric.label,
      };
      groups.push(group);
      groupById.set(group.id, group);
      lastGroupLabels.set(group.id, group.label ?? group.id);
      for (const res of componentIds) assigned.add(res);
    }
  }

  // Try to assign ungrouped resources to their last known existing group.
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

  // Build synthetic groups for cached groups whose anchors disappeared.
  const stillUnassignedIds = Array.from(resourceMap.keys()).filter(
    (id) => !assigned.has(id),
  );
  const syntheticGroups = new Map<string, V1Resource[]>();
  for (const id of stillUnassignedIds) {
    const cachedGroupId = lastGroupAssignments.get(id);
    if (!cachedGroupId) continue;
    if (groupById.has(cachedGroupId)) continue;
    if (!syntheticGroups.has(cachedGroupId)) syntheticGroups.set(cachedGroupId, []);
    const res = resourceMap.get(id);
    if (res) syntheticGroups.get(cachedGroupId)!.push(res);
  }
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
      const rid = toNodeId(res.meta);
      if (rid) assigned.add(rid);
    }
  }

  // Ensure missing-but-cached resources remain in their last group as placeholders.
  const presentIds = new Set(resourceMap.keys());
  for (const [rid, gid] of lastGroupAssignments) {
    // Only if we already have the group in the result
    const grp = groupById.get(gid);
    if (!grp) continue;
    // Already present
    if (presentIds.has(rid)) continue;
    // Avoid duplicate placeholders if already added somehow
    if (grp.resources.some((r) => toNodeId(r.meta) === rid)) continue;
    grp.resources.push(makePlaceholderResource(rid));
    assigned.add(rid);
  }

  // If some cached groups don't exist in this pass (e.g., all members missing),
  // create recovered groups composed of placeholders so the component remains visible.
  for (const [gid, label] of lastGroupLabels) {
    if (groupById.has(gid)) continue;
    const members: V1Resource[] = [];
    for (const [rid, rgid] of lastGroupAssignments) {
      if (rgid !== gid) continue;
      if (assigned.has(rid)) continue; // already included
      members.push(makePlaceholderResource(rid));
      assigned.add(rid);
    }
    if (members.length) {
      const group: ResourceGraphGrouping = { id: gid, resources: members, label };
      groups.push(group);
      groupById.set(gid, group);
    }
  }

  const collectRemainingComponent = (
    startId: string,
    available: Set<string>,
  ) => {
    const queue = [startId];
    const component = new Set<string>();

    while (queue.length) {
      const current = queue.shift()!;
      if (!available.has(current) || component.has(current)) continue;
      component.add(current);

      const neighbors = adjacency.get(current);
      if (!neighbors?.size) continue;
      for (const neighbor of neighbors) {
        if (!available.has(neighbor) || component.has(neighbor)) continue;
        queue.push(neighbor);
      }
    }

    return component;
  };

  const remainingSet = new Set(
    Array.from(resourceMap.keys()).filter((id) => !assigned.has(id)),
  );

  for (const id of unassignedIds) {
    if (!remainingSet.has(id)) continue;
    const componentIds = collectRemainingComponent(id, remainingSet);
    if (!componentIds.size) {
      remainingSet.delete(id);
      continue;
    }

    for (const componentId of componentIds) remainingSet.delete(componentId);

    const componentResources = Array.from(componentIds)
      .map((resourceId) => resourceMap.get(resourceId))
      .filter((res): res is V1Resource => !!res);

    if (componentResources.length) {
      groups.push({
        id,
        resources: componentResources,
        label: "Other resources",
      });
    }
  }

  if (!metricIds.length && !groups.length) {
    const allResources = Array.from(resourceMap.values());
    if (allResources.length) {
      groups.push({
        id: "all-resources",
        resources: allResources,
      });
    }
  }

  // Update caches with current grouping assignments.
  for (const group of groups) {
    if (group.label) lastGroupLabels.set(group.id, group.label);
    for (const res of group.resources) {
      const rid = toNodeId(res.meta);
      if (rid) lastGroupAssignments.set(rid, group.id);
    }
  }

  // Persist grouping cache so we can recover layout/membership on errors
  persistAllCaches();

  return groups;
}
