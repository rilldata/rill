import { writable } from "svelte/store";
import type { ResourceNodeData } from "../shared/types";

/**
 * Store for tracking the currently selected node in the graph inspector.
 * This allows the inspector panel to reactively update when a node is selected.
 */
export const selectedGraphNode = writable<ResourceNodeData | null>(null);

/**
 * Select a node to display in the inspector
 */
export function selectGraphNode(nodeData: ResourceNodeData | null) {
  selectedGraphNode.set(nodeData);
}

/**
 * Clear the current selection
 */
export function clearGraphNodeSelection() {
  selectedGraphNode.set(null);
}
