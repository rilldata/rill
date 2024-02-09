import type {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import { get, type Readable, type Writable } from "svelte/store";

/**
 *  TODO: We should create a single user preference store for all dashboards
 *  and have it in sync with the Cloud user preference store
 *
 *  This store would be similar to MetricsExplorerEntityStore but for user preferences
 */
export interface LocalUserPreferences {
  timeZone?: string;

  visibleMeasures?: string[];
  visibleDimensions?: string[];
  leaderboardMeasureName?: string;

  dashboardSortType?: SortType;
  sortDirection?: SortDirection;
}
let localUserPreferences: Writable<LocalUserPreferences>;

export function initLocalUserPreferenceStore(metricViewName: string) {
  localUserPreferences = localStorageStore<LocalUserPreferences>(
    `${metricViewName}-userPreference`,
    {
      timeZone: getLocalIANA(),
    },
  );

  return localUserPreferences;
}

function localUserPreferencesActions() {
  function updateKey<K extends keyof LocalUserPreferences>(key: K) {
    return (val: LocalUserPreferences[K]) => {
      if (!localUserPreferences) return;
      localUserPreferences.update((lup) => {
        lup[key] = val;
        return lup;
      });
    };
  }

  return {
    updateTimeZone: updateKey("timeZone"),
    updateVisibleMeasures: updateKey("visibleMeasures"),
    updateVisibleDimensions: updateKey("visibleDimensions"),
    updateLeaderboardMeasureName: updateKey("leaderboardMeasureName"),
    updateDashboardSortType: updateKey("dashboardSortType"),
    updateSortDirection: updateKey("sortDirection"),
    reset() {
      // cleanup dashboard settings. note that `timeZone` is not reset.
      // it is intentional because it is an old feature
      localUserPreferences.update((lup) => {
        delete lup.visibleMeasures;
        delete lup.visibleDimensions;
        delete lup.leaderboardMeasureName;
        delete lup.dashboardSortType;
        delete lup.sortDirection;
        return lup;
      });
    },
  };
}

export function getLocalUserPreferences(): Readable<LocalUserPreferences> &
  ReturnType<typeof localUserPreferencesActions> {
  return {
    subscribe: localUserPreferences.subscribe,
    ...localUserPreferencesActions(),
  };
}

export function getLocalUserPreferencesState(): LocalUserPreferences {
  if (!localUserPreferences) return {};
  return get(localUserPreferences);
}
