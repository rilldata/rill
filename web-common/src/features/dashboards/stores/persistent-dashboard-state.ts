import type {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { LocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { get, type Readable, type Updater, type Writable } from "svelte/store";

/**
 * Partial state of the dashboard that is stored in local storage.
 */
export type PersistentDashboardState = {
  visibleMeasures?: string[];
  visibleDimensions?: string[];
  leaderboardMeasureName?: string;

  dashboardSortType?: SortType;
  sortDirection?: SortDirection;
};

function persistentDashboardActions(
  update: (this: void, updater: Updater<PersistentDashboardState>) => void,
) {
  function updateKey<K extends keyof PersistentDashboardState>(key: K) {
    return (val: PersistentDashboardState[K]) => {
      update((lup) => {
        lup[key] = val;
        return lup;
      });
    };
  }

  return {
    updateVisibleMeasures: updateKey("visibleMeasures"),
    updateVisibleDimensions: updateKey("visibleDimensions"),
    updateLeaderboardMeasureName: updateKey("leaderboardMeasureName"),
    updateDashboardSortType: updateKey("dashboardSortType"),
    updateSortDirection: updateKey("sortDirection"),
    reset() {
      // cleanup dashboard settings. note that `timeZone` is not reset.
      // it is intentional because it is an old feature
      update((pd) => {
        delete pd.visibleMeasures;
        delete pd.visibleDimensions;
        delete pd.leaderboardMeasureName;
        delete pd.dashboardSortType;
        delete pd.sortDirection;
        return pd;
      });
    },
  };
}

export type PersistentDashboardStore = Readable<PersistentDashboardState> &
  ReturnType<typeof persistentDashboardActions>;
export function createPersistentDashboardStore(storeKey: string) {
  const { subscribe, update } = localStorageStore<PersistentDashboardState>(
    `${storeKey}-userPreference`,
    {},
  );
  return {
    subscribe,
    ...persistentDashboardActions(update),
  };
}

// TODO: once we move everything to state-managers we wont need this
let persistentDashboardStore: PersistentDashboardStore;
export function initPersistentDashboardStore(storeKey: string) {
  persistentDashboardStore = createPersistentDashboardStore(storeKey);
}

export function getPersistentDashboardStore() {
  return persistentDashboardStore;
}

export function getPersistentDashboardState(): PersistentDashboardState {
  if (!persistentDashboardStore) return {};
  return get(persistentDashboardStore);
}
