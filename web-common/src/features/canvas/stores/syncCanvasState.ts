import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { canvasEntities } from "@rilldata/web-common/features/canvas/stores/canvas-entities";
import { get } from "svelte/store";

export function createCanvasStateSync(ctx: StateManagers) {
  const canvasName = get(ctx.canvasName);
  if (canvasEntities.hasCanvas(canvasName)) {
    // TODO: Add sync method if required
    return { isFetching: false, error: false };
  } else {
    // Running for the 1st time. Initialise the canvas store.
    canvasEntities.addCanvas(canvasName);
    return { isFetching: false, error: false };
  }
}
