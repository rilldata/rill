import { writable } from "svelte/store";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

const anchorStore = writable<V1Resource | null>(null);

export const resourceGraphOverlayAnchor = {
  subscribe: anchorStore.subscribe,
};

export function openResourceGraphOverlay(resource?: V1Resource | null) {
  anchorStore.set(resource ?? null);
}

export function closeResourceGraphOverlay() {
  anchorStore.set(null);
}
