import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import type { Readable, Writable } from "svelte/store";

/**
 *  TODO: We should create a single user preference store for all dashboards
 *  and have it in sync with the Cloud user preference store
 *
 *  This store would be similar to MetricsExplorerEntityStore but for user preferences
 */
export interface LocalUserPreferences {
  timeZone?: string;
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
  };
}

export function getLocalUserPreferences(): Readable<LocalUserPreferences> &
  ReturnType<typeof localUserPreferencesActions> {
  return {
    subscribe: localUserPreferences.subscribe,
    ...localUserPreferencesActions(),
  };
}
