import type { Edge } from "@xyflow/svelte";

/**
 * Result of a graph traversal operation containing visited nodes and edges.
 */
export interface TraversalResult {
  /** Set of node IDs that were visited during traversal */
  visited: Set<string>;
  /** Set of edge IDs that were traversed */
  edgeIds: Set<string>;
}

/**
 * Build adjacency index maps from an edge list for O(1) neighbor lookups.
 * Call once per edge set, then pass the maps to traversal functions.
 */
function buildEdgeIndex(edges: Edge[]): {
  byTarget: Map<string, Edge[]>;
  bySource: Map<string, Edge[]>;
} {
  const byTarget = new Map<string, Edge[]>();
  const bySource = new Map<string, Edge[]>();

  for (const e of edges) {
    const tList = byTarget.get(e.target);
    if (tList) tList.push(e);
    else byTarget.set(e.target, [e]);

    const sList = bySource.get(e.source);
    if (sList) sList.push(e);
    else bySource.set(e.source, [e]);
  }

  return { byTarget, bySource };
}

/**
 * Traverse upstream (sources) from selected nodes via incoming edges only.
 * Uses breadth-first search to find all nodes that the selected nodes depend on.
 *
 * Complexity: O(V + E) — edges are indexed once, then each edge is visited at most once.
 *
 * @param selectedIds - Set of node IDs to start traversal from
 * @param edges - All edges in the graph
 * @returns Object containing visited node IDs and traversed edge IDs
 */
export function traverseUpstream(
  selectedIds: Set<string>,
  edges: Edge[],
): TraversalResult {
  const { byTarget } = buildEdgeIndex(edges);
  const visited = new Set<string>();
  const edgeIds = new Set<string>();
  const queue: string[] = Array.from(selectedIds);
  let queueIndex = 0;

  while (queueIndex < queue.length) {
    const curr = queue[queueIndex++];
    if (visited.has(curr)) continue;
    visited.add(curr);

    const incoming = byTarget.get(curr);
    if (!incoming) continue;
    for (const e of incoming) {
      edgeIds.add(e.id);
      if (!visited.has(e.source)) queue.push(e.source);
    }
  }

  return { visited, edgeIds };
}

/**
 * Traverse downstream (dependents) from selected nodes via outgoing edges only.
 * Uses breadth-first search to find all nodes that depend on the selected nodes.
 *
 * Complexity: O(V + E) — edges are indexed once, then each edge is visited at most once.
 *
 * @param selectedIds - Set of node IDs to start traversal from
 * @param edges - All edges in the graph
 * @returns Object containing visited node IDs and traversed edge IDs
 */
export function traverseDownstream(
  selectedIds: Set<string>,
  edges: Edge[],
): TraversalResult {
  const { bySource } = buildEdgeIndex(edges);
  const visited = new Set<string>();
  const edgeIds = new Set<string>();
  const queue: string[] = Array.from(selectedIds);
  let queueIndex = 0;

  while (queueIndex < queue.length) {
    const curr = queue[queueIndex++];
    if (visited.has(curr)) continue;
    visited.add(curr);

    const outgoing = bySource.get(curr);
    if (!outgoing) continue;
    for (const e of outgoing) {
      edgeIds.add(e.id);
      if (!visited.has(e.target)) queue.push(e.target);
    }
  }

  return { visited, edgeIds };
}
