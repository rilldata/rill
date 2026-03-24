import { writable } from "svelte/store";
import type { ResourceNodeData } from "../shared/types";

export interface InspectState {
  data: ResourceNodeData;
  /** Position of the node relative to the graph container */
  x: number;
  y: number;
  width: number;
  height: number;
}

export const inspectedNode = writable<InspectState | null>(null);

export function openInspect(
  data: ResourceNodeData,
  rect: { x: number; y: number; width: number; height: number },
) {
  inspectedNode.set({ data, ...rect });
}

export function closeInspect() {
  inspectedNode.set(null);
}
