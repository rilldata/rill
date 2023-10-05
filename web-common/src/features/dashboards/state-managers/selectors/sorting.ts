import { derived, type Readable } from "svelte/store";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../../dashboard-stores";

export type SortingSelectors = ReturnType<typeof createSortingSelectors>;

export const createSortingSelectors = (
  dashboardStore: Readable<MetricsExplorerEntity>
) => {
  return {
    sortType: derived(
      dashboardStore,
      (dashboard) => dashboard.dashboardSortType
    ),
    sortedAscending: derived(
      dashboardStore,
      (dashboard) => dashboard.sortDirection === SortDirection.ASCENDING
    ),
    sortMeasure: derived(dashboardStore, (dashboard) =>
      dashboard.dashboardSortType !== SortType.DIMENSION &&
      dashboard.dashboardSortType !== SortType.UNSPECIFIED
        ? dashboard.leaderboardMeasureName
        : null
    ),
    sortedByDimensionValue: derived(
      dashboardStore,
      (dashboard) => dashboard.dashboardSortType === SortType.DIMENSION
    ),
  };
};
