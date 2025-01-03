import { canvasEntities } from "@rilldata/web-common/features/canvas/stores/canvas-entities";

export function createCanvasStateSync(canvasName: string) {
  if (canvasEntities.hasCanvas(canvasName)) {
    // TODO: Add sync method if required
    return { isFetching: false, error: false };
  } else {
    // Running for the 1st time. Initialise the canvas store.
    canvasEntities.addCanvas(canvasName);
    return { isFetching: false, error: false };
  }
}
