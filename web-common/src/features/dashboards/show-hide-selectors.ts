import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
import {
  updateMetricsExplorerByName,
  useDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  MetricsViewDimension,
  MetricsViewSpecMeasureV2,
  RpcStatus,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { Readable, derived, get } from "svelte/store";

export type ShowHideSelectorState = {
  selectableItems: Array<SearchableFilterSelectableItem>;
  selectedItems: Array<boolean>;
  availableKeys: Array<string>;
};
type ShowHideSelectorReducers = {
  toggleVisibility: (key: string) => void;
  setAllToVisible: () => void;
  setAllToNotVisible: () => void;
};
export type ShowHideSelectorStore = Readable<ShowHideSelectorState> &
  ShowHideSelectorReducers;

function createShowHideStore<Item>(
  metricsViewName: string,
  metricsView: CreateQueryResult<V1MetricsView, RpcStatus>,
  fieldInMeta: keyof Pick<V1MetricsView, "dimensions" | "measures">,
  visibleFieldInStore: keyof Pick<
    MetricsExplorerEntity,
    "visibleMeasureKeys" | "visibleDimensionKeys"
  >,
  allVisibleFieldInStore: keyof Pick<
    MetricsExplorerEntity,
    "allMeasuresVisible" | "allDimensionsVisible"
  >,
  labelSelector: (i: Item) => string,
) {
  const derivedStore = derived(
    [metricsView, useDashboardStore(metricsViewName)],
    ([meta, metricsExplorer]) => {
      if (
        !meta?.data ||
        !metricsExplorer ||
        !meta.isSuccess ||
        meta.isRefetching
      ) {
        return {
          selectableItems: [],
          selectedItems: [],
          availableKeys: [],
        };
      }

      const items = meta.data[fieldInMeta];
      const selectableItems: Array<SearchableFilterSelectableItem> = items.map(
        (i) => ({
          name: i.name,
          label: labelSelector(i),
        }),
      );
      const availableKeys = items.map((i) => i.name);
      const visibleKeysSet = metricsExplorer[visibleFieldInStore];

      return {
        selectableItems,
        selectedItems: availableKeys.map((k) => visibleKeysSet.has(k)),
        availableKeys,
      };
    },
  ) as ShowHideSelectorStore;

  derivedStore.setAllToVisible = () => {
    updateMetricsExplorerByName(metricsViewName, (metricsExplorer) => {
      metricsExplorer[visibleFieldInStore] = new Set(
        get(derivedStore).availableKeys,
      );
      metricsExplorer[allVisibleFieldInStore] = true;
    });
  };

  derivedStore.setAllToNotVisible = () => {
    updateMetricsExplorerByName(metricsViewName, (metricsExplorer) => {
      // Remove all keys except for the first one
      const firstKey = get(derivedStore).availableKeys.slice(0, 1);
      metricsExplorer[visibleFieldInStore] = new Set(firstKey);

      if (fieldInMeta === "measures") {
        metricsExplorer.leaderboardMeasureName = firstKey[0];
      }
      metricsExplorer[allVisibleFieldInStore] = false;
    });
  };

  derivedStore.toggleVisibility = (key) => {
    updateMetricsExplorerByName(metricsViewName, (metricsExplorer) => {
      if (metricsExplorer[visibleFieldInStore].has(key)) {
        metricsExplorer[visibleFieldInStore].delete(key);

        /*
         * If current leaderboard measure is hidden, set the first
         * visible measure as the current leaderboard measure
         */
        if (
          fieldInMeta === "measures" &&
          metricsExplorer.leaderboardMeasureName === key
        ) {
          /*
           * To maintain the order of keys, filter out the
           * non-visible ones from the available keys
           */
          const firstVisible = get(derivedStore).availableKeys.find((key) =>
            metricsExplorer[visibleFieldInStore].has(key),
          );

          metricsExplorer.leaderboardMeasureName = firstVisible;
        }
      } else {
        metricsExplorer[visibleFieldInStore].add(key);
      }
      metricsExplorer[allVisibleFieldInStore] =
        metricsExplorer[visibleFieldInStore].size ===
        get(derivedStore).availableKeys.length;
    });
  };

  return derivedStore;
}

export function createShowHideMeasuresStore(
  metricsViewName: string,
  metricsView: CreateQueryResult<V1MetricsView, RpcStatus>,
) {
  return createShowHideStore<MetricsViewSpecMeasureV2>(
    metricsViewName,
    metricsView,
    "measures",
    "visibleMeasureKeys",
    "allMeasuresVisible",
    /*
     * This selector returns the best available string for each measure,
     * using the "label" if available but falling back to the expression
     * if needed.
     */
    (m) => m.label || m.expression,
  );
}

export function createShowHideDimensionsStore(
  metricsViewName: string,
  metricsView: CreateQueryResult<V1MetricsView, RpcStatus>,
) {
  return createShowHideStore<MetricsViewDimension>(
    metricsViewName,
    metricsView,
    "dimensions",
    "visibleDimensionKeys",
    "allDimensionsVisible",
    /*
     * This selector returns the best available string for each dimension,
     * using the "label" if available but falling back to the name of
     * the categorical column (which must be present) if needed
     */
    (d) => d.label || d.name,
  );
}
