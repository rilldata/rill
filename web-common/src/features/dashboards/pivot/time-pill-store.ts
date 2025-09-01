import {
  getAllowedTimeGrains,
  isAvailableTimeGrain,
  isGrainBigger,
} from "@rilldata/web-common/lib/time/grains";
import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { derived, writable } from "svelte/store";
import type { PivotChipData } from "./types";
import { PivotChipType } from "./types";

export interface TimePillState {
  label: string;
  usedGrains: AvailableTimeGrain[];
  availableGrains: AvailableTimeGrain[];
  allGrainsUsed: boolean;
}

export interface TimePillStoreState {
  [timeDimension: string]: TimePillState;
}

export interface TimeControlsInfo {
  timeStart: string;
  timeEnd: string;
  minTimeGrain?: V1TimeGrain;
}

const timePillStore = writable<TimePillStoreState>({});

const timeControlsInfo = writable<TimeControlsInfo | null>(null);

export const timePills = derived(
  [timePillStore, timeControlsInfo],
  ([pillState, timeControls]) => {
    if (!timeControls) return pillState;

    const updatedState: TimePillStoreState = {};

    // Update each time dimension's computed properties
    Object.keys(pillState).forEach((timeDimensionKey) => {
      const state = pillState[timeDimensionKey];

      // Get all allowed grains for the time range
      const allAllowedGrains = getAllowedTimeGrains(
        new Date(timeControls.timeStart),
        new Date(timeControls.timeEnd),
      )
        .map((tg) => tg.grain)
        .filter(isAvailableTimeGrain);

      // Filter out grains that are smaller than minTimeGrain
      const validGrains = allAllowedGrains.filter((grain) => {
        if (
          !timeControls.minTimeGrain ||
          timeControls.minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED
        ) {
          return true;
        }
        return !isGrainBigger(timeControls.minTimeGrain, grain);
      });

      const availableGrains = validGrains.filter(
        (grain) => !state.usedGrains.includes(grain),
      );

      const allGrainsUsed = availableGrains.every((grain) =>
        state.usedGrains.includes(grain),
      );

      updatedState[timeDimensionKey] = {
        ...state,
        availableGrains,
        allGrainsUsed,
      };
    });

    return updatedState;
  },
);

export const timePillActions = {
  setTimeControls(
    timeStart: string,
    timeEnd: string,
    minTimeGrain?: V1TimeGrain,
  ) {
    timeControlsInfo.set({ timeStart, timeEnd, minTimeGrain });
  },

  // Initialize a time dimension
  initTimeDimension(timeDimension: string, label: string) {
    timePillStore.update((state) => {
      if (state[timeDimension]) {
        return state;
      }

      return {
        ...state,
        [timeDimension]: {
          label,
          usedGrains: [],
          availableGrains: [],
          allGrainsUsed: false,
        },
      };
    });
  },

  // Update used grains from pivot chips
  updateUsedGrains(
    timeDimensionKey: string,
    rows: PivotChipData[],
    columns: PivotChipData[],
  ) {
    const allTimeChips = [...rows, ...columns].filter(
      (chip) => chip.type === PivotChipType.Time,
    );
    const usedGrains = allTimeChips.map(
      (chip) => chip.id as AvailableTimeGrain,
    );

    timePillStore.update((state) => ({
      ...state,
      [timeDimensionKey]: {
        ...state[timeDimensionKey],
        usedGrains,
      },
    }));
  },
};

export const timePillSelectors = {
  getAvailableGrains: (timeDimensionKey: string) =>
    derived(
      timePills,
      ($store) => $store[timeDimensionKey]?.availableGrains || [],
    ),
  getAllGrainsUsed: (timeDimensionKey: string) =>
    derived(
      timePills,
      ($store) => $store[timeDimensionKey]?.allGrainsUsed || false,
    ),
};
