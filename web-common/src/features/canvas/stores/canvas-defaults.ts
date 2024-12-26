import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export function getDefaultCanvasEntity(name: string): CanvasEntity {
  return {
    name,
    selectedComponentIndex: null,
    selectedTimezone: "UTC",
    selectedTimeRange: {
      name: TimeRangePreset.ALL_TIME,
      start: new Date(0),
      end: new Date(),
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    },
    selectedComparisonTimeRange: undefined,
    showTimeComparison: false,
  };
}

export function restorePersistedCanvasState(canvas: CanvasEntity) {
  // TODO: Implement persistence logic when needed
  return canvas;
}
