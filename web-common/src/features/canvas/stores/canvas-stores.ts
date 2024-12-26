import { getDefaultCanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-defaults";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { derived, writable, type Readable } from "svelte/store";

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

  selectTimeRange(
    name: string,
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    updateCanvasByName(name, (canvas) => {
      if (!timeRange.name) return;

      if (timeRange.name === TimeRangePreset.ALL_TIME) {
        canvas.showTimeComparison = false;
      }

      canvas.selectedTimeRange = {
        ...timeRange,
        interval: timeGrain,
      };

      canvas.selectedComparisonTimeRange = comparisonTimeRange;
    });
  },

  setSelectedComparisonRange(
    name: string,
    comparisonTimeRange: DashboardTimeControls,
  ) {
    updateCanvasByName(name, (canvas) => {
      canvas.selectedComparisonTimeRange = comparisonTimeRange;
    });
  },

  // Update the selected timezone
  setTimeZone(name: string, timezone: string) {
    updateCanvasByName(name, (canvas) => {
      canvas.selectedTimezone = timezone;
    });
  },

  displayTimeComparison(name: string, showTimeComparison: boolean) {
    updateCanvasByName(name, (canvas) => {
      canvas.showTimeComparison = showTimeComparison;
    });
  },
};

export const canvasEntityStore: Readable<CanvasStoreType> &
  typeof canvasVariableReducers = {
  subscribe,
  ...canvasVariableReducers,
};

export function useCanvasStore(name: string): Readable<CanvasEntity> {
  return derived(canvasEntityStore, ($store) => {
    return $store.entities[name];
  });
}
