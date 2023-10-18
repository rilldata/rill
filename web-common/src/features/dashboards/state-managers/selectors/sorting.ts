import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { SelectorFnArgs } from "./types";

export const sortingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  sortType: ({ dashboard }: SelectorFnArgs) => dashboard.dashboardSortType,

  /**
   * true if the dashboard is sorted ascending, false otherwise.
   */
  sortedAscending: ({ dashboard }: SelectorFnArgs) =>
    dashboard.sortDirection === SortDirection.ASCENDING,

  /**
   * Returns the measure name that the dashboard is sorted by,
   * or null if the dashboard is sorted by dimension value.
   */
  sortMeasure: ({ dashboard }: SelectorFnArgs) =>
    dashboard.dashboardSortType !== SortType.DIMENSION &&
    dashboard.dashboardSortType !== SortType.UNSPECIFIED
      ? dashboard.leaderboardMeasureName
      : null,

  /**
   * Returns true if the dashboard is sorted by a dimension, false otherwise.
   */
  sortedByDimensionValue: ({ dashboard }: SelectorFnArgs) =>
    dashboard.dashboardSortType === SortType.DIMENSION,
};
