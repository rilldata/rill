import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { canvasEntityStore } from "@rilldata/web-common/features/canvas/stores/canvas-stores";
import { get } from "svelte/store";

export function createCanvasStateSync(ctx: StateManagers) {
  const canvasName = get(ctx.canvasName);
  if (canvasName in get(canvasEntityStore).entities) {
    // TODO: Add sync method if required
    return { isFetching: false, error: false };
  } else {
    // Running for the 1st time. Initialise the canvas store.
    canvasEntityStore.init(canvasName);
    return { isFetching: false, error: false };
  }
}
