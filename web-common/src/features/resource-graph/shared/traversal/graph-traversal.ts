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
 * Traverse upstream (sources) from selected nodes via incoming edges only.
 * Uses breadth-first search to find all nodes that the selected nodes depend on.
 *
 * @param selectedIds - Set of node IDs to start traversal from
 * @param edges - All edges in the graph
 * @returns Object containing visited node IDs and traversed edge IDs
 *
 * @example
 * const result = traverseUpstream(new Set(['model:orders']), edges);
 * // result.visited contains 'model:orders' and all its upstream sources
 * // result.edgeIds contains all edges leading to 'model:orders'
 */
export function traverseUpstream(
  selectedIds: Set<string>,
  edges: Edge[],
): TraversalResult {
  const visited = new Set<string>();
  const edgeIds = new Set<string>();
  const queue: string[] = Array.from(selectedIds);
  let queueIndex = 0;

  while (queueIndex < queue.length) {
    const curr = queue[queueIndex++];
    if (visited.has(curr)) continue;
    visited.add(curr);

    // Find all incoming edges (edges where current node is the target)
    for (const e of edges) {
      if (e.target === curr) {
        edgeIds.add(e.id);
        if (!visited.has(e.source)) queue.push(e.source);
      }
    }
  }

  return { visited, edgeIds };
}

/**
 * Traverse downstream (dependents) from selected nodes via outgoing edges only.
 * Uses breadth-first search to find all nodes that depend on the selected nodes.
 *
 * @param selectedIds - Set of node IDs to start traversal from
 * @param edges - All edges in the graph
 * @returns Object containing visited node IDs and traversed edge IDs
 *
 * @example
 * const result = traverseDownstream(new Set(['source:users']), edges);
 * // result.visited contains 'source:users' and all models/metrics that use it
 * // result.edgeIds contains all edges starting from 'source:users'
 */
export function traverseDownstream(
  selectedIds: Set<string>,
  edges: Edge[],
): TraversalResult {
  const visited = new Set<string>();
  const edgeIds = new Set<string>();
  const queue: string[] = Array.from(selectedIds);
  let queueIndex = 0;

  while (queueIndex < queue.length) {
    const curr = queue[queueIndex++];
    if (visited.has(curr)) continue;
    visited.add(curr);

    // Find all outgoing edges (edges where current node is the source)
    for (const e of edges) {
      if (e.source === curr) {
        edgeIds.add(e.id);
        if (!visited.has(e.target)) queue.push(e.target);
      }
    }
  }

  return { visited, edgeIds };
}

/**
 * Traverse both upstream and downstream from selected nodes.
 * Combines the results of both traversal directions.
 *
 * @param selectedIds - Set of node IDs to start traversal from
 * @param edges - All edges in the graph
 * @returns Object containing all visited node IDs and traversed edge IDs
 *
 * @example
 * const result = traverseBidirectional(new Set(['model:orders']), edges);
 * // result.visited contains all nodes connected to 'model:orders'
 * // result.edgeIds contains all edges in the connected component
 */
export function traverseBidirectional(
  selectedIds: Set<string>,
  edges: Edge[],
): TraversalResult {
  const upstream = traverseUpstream(selectedIds, edges);
  const downstream = traverseDownstream(selectedIds, edges);

  return {
    visited: new Set([...upstream.visited, ...downstream.visited]),
    edgeIds: new Set([...upstream.edgeIds, ...downstream.edgeIds]),
  };
}
