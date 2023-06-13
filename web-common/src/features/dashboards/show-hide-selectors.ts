import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
import {
  MetricsExplorerEntity,
  updateMetricsExplorerByName,
  useDashboardStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type {
  MetricsViewDimension,
  MetricsViewMeasure,
  RpcStatus,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, Readable } from "svelte/store";

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
  metaQuery: CreateQueryResult<V1MetricsView, RpcStatus>,
  fieldInMeta: keyof Pick<V1MetricsView, "dimensions" | "measures">,
  visibleFieldInStore: keyof Pick<
    MetricsExplorerEntity,
    "visibleMeasureKeys" | "visibleDimensionKeys"
  >,
  labelSelector: (i: Item) => string
) {
  const derivedStore = derived(
    [metaQuery, useDashboardStore(metricsViewName)],
    ([meta, metricsExplorer]) => {
      if (!meta || !meta.isSuccess || meta.isRefetching) {
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
        })
      );
      const availableKeys = items.map((i) => i.name);
      const visibleKeysSet = metricsExplorer[visibleFieldInStore];

      return {
        selectableItems,
        selectedItems: availableKeys.map((k) => visibleKeysSet.has(k)),
        availableKeys,
      };
    }
  ) as ShowHideSelectorStore;

  derivedStore.setAllToVisible = () => {
    updateMetricsExplorerByName(metricsViewName, (metricsExplorer) => {
      metricsExplorer[visibleFieldInStore] = new Set(
        get(derivedStore).availableKeys
      );
    });
  };

  derivedStore.setAllToNotVisible = () => {
    updateMetricsExplorerByName(metricsViewName, (metricsExplorer) => {
      metricsExplorer[visibleFieldInStore].clear();
    });
  };

  derivedStore.toggleVisibility = (key) => {
    updateMetricsExplorerByName(metricsViewName, (metricsExplorer) => {
      if (metricsExplorer[visibleFieldInStore].has(key)) {
        metricsExplorer[visibleFieldInStore].delete(key);
      } else {
        metricsExplorer[visibleFieldInStore].add(key);
      }
    });
  };

  return derivedStore;
}

export function createShowHideMeasuresStore(
  metricsViewName: string,
  metaQuery: CreateQueryResult<V1MetricsView, RpcStatus>
) {
  return createShowHideStore<MetricsViewMeasure>(
    metricsViewName,
    metaQuery,
    "measures",
    "visibleMeasureKeys",
    (m) => m.label || m.expression
  );
}

export function createShowHideDimensionsStore(
  metricsViewName: string,
  metaQuery: CreateQueryResult<V1MetricsView, RpcStatus>
) {
  return createShowHideStore<MetricsViewDimension>(
    metricsViewName,
    metaQuery,
    "dimensions",
    "visibleDimensionKeys",
    (d) => d.label || d.name
  );
}
