import { getContext, setContext } from "svelte";
import { writable, type Writable } from "svelte/store";
import type { ResourceNodeData } from "../shared/types";

export interface InspectState {
  data: ResourceNodeData;
  /** Position of the node relative to the graph container */
  x: number;
  y: number;
  width: number;
  height: number;
}

const INSPECT_CONTEXT_KEY = Symbol("inspect-store");

/**
 * Create a new inspect store and set it in the component's Svelte context.
 * Call this in each GraphCanvas instance so each canvas gets its own store.
 */
export function initInspectStore(): Writable<InspectState | null> {
  const store = writable<InspectState | null>(null);
  setContext(INSPECT_CONTEXT_KEY, store);
  return store;
}

/**
 * Retrieve the inspect store from Svelte context.
 * Falls back to a standalone store if no context is available (e.g., tests).
 */
export function getInspectStore(): Writable<InspectState | null> {
  try {
    return getContext<Writable<InspectState | null>>(INSPECT_CONTEXT_KEY);
  } catch {
    // Outside a component lifecycle (tests); return a fresh store.
    return writable<InspectState | null>(null);
  }
}

export function openInspect(
  store: Writable<InspectState | null>,
  data: ResourceNodeData,
  rect: { x: number; y: number; width: number; height: number },
) {
  store.set({ data, ...rect });
}

export function closeInspect(store: Writable<InspectState | null>) {
  store.set(null);
}
