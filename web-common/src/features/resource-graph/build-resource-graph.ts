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

function toNodeId(meta?: V1ResourceMeta): string | undefined {
  if (!meta?.name?.name || !meta?.name?.kind) return undefined;
  return `${meta.name.kind}:${meta.name.name}`;
}

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
    nodesep: 120,
    ranksep: 160,
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
    node.position = {
      x: dagreNode.x - nodeWidth / 2,
      y: dagreNode.y - (node.height ?? DEFAULT_NODE_HEIGHT) / 2,
    };
    node.targetPosition = Position.Top;
    node.sourcePosition = Position.Bottom;
  }

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
      groups.push({
        id: metric.id,
        resources: componentResources,
        label: metric.label,
      });
      for (const res of componentIds) assigned.add(res);
    }
  }

  const unassignedIds = Array.from(resourceMap.keys()).filter(
    (id) => !assigned.has(id),
  );

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

  const remainingSet = new Set(unassignedIds);

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

  return groups;
}
