import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";

export function getDefaultCanvasEntity(name: string): CanvasEntity {
  return {
    name,
    selectedComponentIndex: null,
    selectedTimezone: "UTC",
    selectedTimeRange: undefined,
  };
}

export function restorePersistedCanvasState(canvas: CanvasEntity) {
  // TODO: Implement persistence logic when needed
  return canvas;
}
