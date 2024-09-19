import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
import {
  updateMetricsExplorerByName,
  useDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getPersistentDashboardStore } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { ValidExploreResponse } from "@rilldata/web-common/features/explores/selectors";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  RpcStatus,
} from "@rilldata/web-common/runtime-client";
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
  exploreName: string,
  validSpecStore: Readable<
    QueryObserverResult<ValidExploreResponse, RpcStatus>
  >,
  type: "dimensions" | "measures",
  labelSelector: (i: Item) => string,
) {
  const typeInCaps = type.replace(/\w/, (s) => s.toUpperCase());
  const visibleFieldInStore = `visible${typeInCaps.slice(
    0,
    typeInCaps.length - 1,
  )}Keys` as keyof Pick<
    MetricsExplorerEntity,
    "visibleMeasureKeys" | "visibleDimensionKeys"
  >;
  const allVisibleFieldInStore = `all${typeInCaps}Visible` as keyof Pick<
    MetricsExplorerEntity,
    "allMeasuresVisible" | "allDimensionsVisible"
  >;
  const persistenceStoreUpdateMethod = `updateVisible${typeInCaps}` as
    | "updateVisibleMeasures"
    | "updateVisibleDimensions";
  const persistentStore = getPersistentDashboardStore();

  const derivedStore = derived(
    [validSpecStore, useDashboardStore(exploreName)],
    ([validSpec, metricsExplorer]) => {
      if (
        !validSpec?.data?.metricsView ||
        !validSpec?.data?.explore ||
        !metricsExplorer ||
        !validSpec.isSuccess ||
        validSpec.isRefetching
      ) {
        return {
          selectableItems: [],
          selectedItems: [],
          availableKeys: [],
        };
      }

      const items = validSpec.data.explore[type] ?? [];
      const selectableItems: Array<SearchableFilterSelectableItem> = items.map(
        (i) => ({
          name: i,
          label: labelSelector(
            validSpec.data.metricsView?.[type]?.find(
              (r) => r.name === i,
            ) as Item,
          ),
        }),
      );
      const availableKeys = [...items];
      const visibleKeysSet = metricsExplorer[visibleFieldInStore];

      return {
        selectableItems,
        selectedItems: availableKeys.map((k) => visibleKeysSet.has(k)),
        availableKeys,
      };
    },
  ) as ShowHideSelectorStore;

  derivedStore.setAllToVisible = () => {
    updateMetricsExplorerByName(exploreName, (metricsExplorer) => {
      metricsExplorer[visibleFieldInStore] = new Set(
        get(derivedStore).availableKeys,
      );
      metricsExplorer[allVisibleFieldInStore] = true;
      persistentStore[persistenceStoreUpdateMethod]([
        ...metricsExplorer[visibleFieldInStore].keys(),
      ]);
    });
  };

  derivedStore.setAllToNotVisible = () => {
    updateMetricsExplorerByName(exploreName, (metricsExplorer) => {
      // Remove all keys except for the first one
      const firstKey = get(derivedStore).availableKeys.slice(0, 1);
      metricsExplorer[visibleFieldInStore] = new Set(firstKey);
      persistentStore[persistenceStoreUpdateMethod]([
        ...metricsExplorer[visibleFieldInStore].keys(),
      ]);

      metricsExplorer[allVisibleFieldInStore] = false;
    });
  };

  derivedStore.toggleVisibility = (key) => {
    updateMetricsExplorerByName(exploreName, (metricsExplorer) => {
      if (metricsExplorer[visibleFieldInStore].has(key)) {
        metricsExplorer[visibleFieldInStore].delete(key);
      } else {
        metricsExplorer[visibleFieldInStore].add(key);
      }
      metricsExplorer[allVisibleFieldInStore] =
        metricsExplorer[visibleFieldInStore].size ===
        get(derivedStore).availableKeys.length;
      persistentStore[persistenceStoreUpdateMethod]([
        ...metricsExplorer[visibleFieldInStore].keys(),
      ]);
    });
  };

  return derivedStore;
}

export function createShowHideMeasuresStore(
  exploreName: string,
  validSpecStore: Readable<
    QueryObserverResult<ValidExploreResponse, RpcStatus>
  >,
) {
  return createShowHideStore<MetricsViewSpecMeasureV2>(
    exploreName,
    validSpecStore,
    "measures",
    /*
     * This selector returns the best available string for each measure,
     * using the "label" if available but falling back to the expression
     * if needed.
     */
    (m) => m.label || m.expression || m.name!,
  );
}

export function createShowHideDimensionsStore(
  exploreName: string,
  validSpecStore: Readable<
    QueryObserverResult<ValidExploreResponse, RpcStatus>
  >,
) {
  return createShowHideStore<MetricsViewSpecDimensionV2>(
    exploreName,
    validSpecStore,
    "dimensions",
    /*
     * This selector returns the best available string for each dimension,
     * using the "label" if available but falling back to the name of
     * the categorical column (which must be present) if needed
     */
    (d) => d.label || d.name!,
  );
}
