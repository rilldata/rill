import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

function getKeyForLocalStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  return `rill:app:explore:${storageNamespacePrefix ?? ""}${exploreName}`.toLowerCase();
}

export function getMostRecentExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
) {
  try {
    const rawUrlSearch = localStorage.getItem(
      getKeyForLocalStore(exploreName, storageNamespacePrefix),
    );
    if (!rawUrlSearch) return { partialExploreState: undefined, errors: [] };

    return convertURLSearchParamsToExploreState(
      new URLSearchParams(rawUrlSearch),
      metricsView,
      explore,
      // Send empty preset so that fields are always stored.
      {},
    );
  } catch {
    // no-op
  }
  return { partialExploreState: undefined, errors: [] };
}

export function saveMostRecentExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  explore: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
  exploreState: MetricsExplorerEntity,
) {
  const urlSearchParams = convertExploreStateToURLSearchParams(
    exploreState,
    explore,
    timeControlsState,
    {},
  );

  try {
    setMostRecentExploreState(
      exploreName,
      storageNamespacePrefix,
      urlSearchParams.toString(),
    );
  } catch {
    // no-op
  }
}

export function clearMostRecentExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  const key = getKeyForLocalStore(exploreName, storageNamespacePrefix);
  localStorage.removeItem(key);
}

export function setMostRecentExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  state: string,
) {
  localStorage.setItem(
    getKeyForLocalStore(exploreName, storageNamespacePrefix),
    state,
  );
}
