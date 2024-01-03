import { removeIfExists } from "@rilldata/web-common/lib/arrayUtils";
import type { DashboardMutables } from "./types";
import { filtersForCurrentExcludeMode } from "../selectors/dimension-filters";
import { potentialFilterName } from "../../filters/Filters.svelte";

export function toggleDimensionValueSelection(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string,
  dimensionValue: string,
  keepPillVisible?: boolean
) {
  const filters = filtersForCurrentExcludeMode({ dashboard })(dimensionName);
  // if there are no filters at this point we cannot update anything.
  if (filters === undefined) {
    return;
  }

  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const dimensionEntryIndex =
    filters.findIndex((filter) => filter.name === dimensionName) ?? -1;

  if (dimensionEntryIndex >= 0) {
    const filtersIn = filters[dimensionEntryIndex].in;
    if (filtersIn === undefined) return;
    if (removeIfExists(filtersIn, (value) => value === dimensionValue)) {
      if (filtersIn.length === 0) {
        filters.splice(dimensionEntryIndex, 1);
        if (keepPillVisible) potentialFilterName.set(dimensionName);
      }
      return;
    }
    filtersIn.push(dimensionValue);
  } else {
    potentialFilterName.set(null);

    filters.push({
      name: dimensionName,
      in: [dimensionValue],
    });
  }
}

export function toggleDimensionNameSelection(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string
) {
  const filters = filtersForCurrentExcludeMode({ dashboard })(dimensionName);
  // if there are no filters at this point we cannot update anything.
  if (filters === undefined) {
    return;
  }

  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const filterIndex = filters.findIndex(
    (filter) => filter.name === dimensionName
  );

  if (filterIndex === -1) {
    filters.push({
      name: dimensionName,
      in: [],
    });
  } else {
    filters.splice(filterIndex, 1);
  }
}

export const dimensionFilterActions = {
  /**
   * Toggles whether the given dimension value is selected in the
   * dimension filter for the given dimension.
   *
   * Note that this is different than the include/exclude mode for
   * dimension filters. This is a toggle for a specific value, whereas
   * the include/exclude mode is a toggle for the entire dimension.
   */
  toggleDimensionValueSelection,
  toggleDimensionNameSelection,
};
