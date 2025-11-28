import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { writable } from "svelte/store";

type QuickViewState = {
  open: boolean;
  anchorResource: V1Resource | null;
};

const initialState: QuickViewState = {
  open: false,
  anchorResource: null,
};

export const resourceGraphQuickViewState =
  writable<QuickViewState>(initialState);

export function openResourceGraphQuickView(resource?: V1Resource | null) {
  if (!resource) {
    console.warn(
      "[ResourceGraphQuickView] Cannot open resource graph without a resource.",
    );
    return;
  }
  resourceGraphQuickViewState.set({
    open: true,
    anchorResource: resource,
  });
}

export function closeResourceGraphQuickView() {
  resourceGraphQuickViewState.set(initialState);
}
