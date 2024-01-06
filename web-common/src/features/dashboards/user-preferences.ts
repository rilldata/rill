import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import type { Writable } from "svelte/store";

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

export function getLocalUserPreferences() {
  return localUserPreferences;
}
