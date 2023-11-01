import { removeIfExists } from "@rilldata/web-common/lib/arrayUtils";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export function toggleDimensionFilter(
  metricsExplorer: MetricsExplorerEntity,
  dimensionName: string,
  dimensionValue: string
) {
  const relevantFilterKey = metricsExplorer.dimensionFilterExcludeMode.get(
    dimensionName
  )
    ? "exclude"
    : "include";

  const filters = metricsExplorer?.filters[relevantFilterKey];

  // if there are no filters at this point we cannot update anything.
  if (filters === undefined) {
    return;
  }

  const dimensionEntryIndex =
    filters.findIndex((filter) => filter.name === dimensionName) ?? -1;

  if (dimensionEntryIndex >= 0) {
    const filtersIn = filters[dimensionEntryIndex].in;
    if (filtersIn === undefined) return;
    if (removeIfExists(filtersIn, (value) => value === dimensionValue)) {
      if (filtersIn.length === 0) {
        filters.splice(dimensionEntryIndex, 1);
      }
      return;
    }
    filtersIn.push(dimensionValue);
  } else {
    filters.push({
      name: dimensionName,
      in: [dimensionValue],
    });
  }
}

export const dimensionFilterActions = {
  /**
   * Toggles the filter of the given dimension value.
   */
  toggleDimensionFilter,
};
