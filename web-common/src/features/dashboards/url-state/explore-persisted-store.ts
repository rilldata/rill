import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  convertPresetToExploreState,
  CustomTimeRangeRegex,
} from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

type PersistedExploreKey<Key extends keyof V1ExplorePreset> = {
  key: Key;
  filter?: (value: V1ExplorePreset[Key]) => boolean;
};

const PersistedExploreKeys: PersistedExploreKey<keyof V1ExplorePreset>[] = [
  <PersistedExploreKey<"timeRange">>{
    key: "timeRange",
    filter: (value) => value && CustomTimeRangeRegex.test(value),
  },
  {
    key: "compareTimeRange",
  },
  {
    key: "measures",
  },
  {
    key: "dimensions",
  },
  {
    key: "exploreSortBy",
  },
  {
    key: "exploreSortAsc",
  },
  {
    key: "exploreSortType",
  },
];

function getKeyForLocalStore(exploreName: string, prefix: string | undefined) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}`.toLowerCase();
}

export function getExploreStateFromLocalStorage(
  exploreName: string,
  prefix: string | undefined,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
): Partial<MetricsExplorerEntity> | undefined {
  try {
    const rawExplorePreset = localStorage.getItem(
      getKeyForLocalStore(exploreName, prefix),
    );
    if (!rawExplorePreset) return undefined;
    const parsedExplorePreset = JSON.parse(rawExplorePreset) as V1ExplorePreset;
    const { partialExploreState } = convertPresetToExploreState(
      metricsView,
      explore,
      parsedExplorePreset,
    );
    return partialExploreState;
  } catch {
    // no-op
  }

  return undefined;
}

export function saveExploreStateToLocalStorage() {}
