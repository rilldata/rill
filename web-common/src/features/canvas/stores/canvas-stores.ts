import { getDefaultCanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-defaults";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { GridStack } from "gridstack";
import { derived, type Readable, writable } from "svelte/store";

export interface CanvasStoreType {
  entities: Record<string, CanvasEntity>;
}
const { update, subscribe } = writable({
  entities: {},
} as CanvasStoreType);

export const updateCanvasByName = (
  name: string,
  callback: (canvas: CanvasEntity) => void,
) => {
  update((state) => {
    if (!state.entities[name]) {
      return state;
    }

    callback(state.entities[name]);
    return state;
  });
};

const canvasVariableReducers = {
  init(name: string) {
    update((state) => {
      if (state.entities[name]) return state;

      state.entities[name] = getDefaultCanvasEntity(name);

      return state;
    });
  },

  remove(name: string) {
    update((state) => {
      delete state.entities[name];
      return state;
    });
  },

  // Update the selected timezone
  setSelectedComponentIndex(name: string, index: number | null) {
    updateCanvasByName(name, (canvas) => {
      canvas.selectedComponentIndex = index;
    });
  },

  // Update the selected time range
  setTimeRange(name: string, timeRange: DashboardTimeControls | undefined) {
    updateCanvasByName(name, (canvas) => {
      canvas.selectedTimeRange = timeRange;
    });
  },

  // Update the selected timezone
  setTimezone(name: string, timezone: string) {
    updateCanvasByName(name, (canvas) => {
      canvas.selectedTimezone = timezone;
    });
  },

  setGridstack(name: string, grid: GridStack | null) {
    updateCanvasByName(name, (canvas) => {
      console.log("setting gridstack", name, grid);
      canvas.gridstack = grid;
    });
  },
};

export const canvasStore: Readable<CanvasStoreType> &
  typeof canvasVariableReducers = {
  subscribe,
  ...canvasVariableReducers,
};

export function useCanvasStore(name: string): Readable<CanvasEntity> {
  return derived(canvasStore, ($store) => {
    return $store.entities[name];
  });
}
