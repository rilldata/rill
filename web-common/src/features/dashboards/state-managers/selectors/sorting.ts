import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardDataSources } from "./types";

export const DashboardKeysInSorting: Array<keyof MetricsExplorerEntity> = [
  "dashboardSortType",
  "sortDirection",
  "leaderboardMeasureName",
];
export type SortingDashboard = Pick<
  MetricsExplorerEntity,
  "dashboardSortType" | "sortDirection" | "leaderboardMeasureName"
>;

export const sortingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  sortType: ({ dashboard }: DashboardDataSources<SortingDashboard>) =>
    dashboard.dashboardSortType,

  /**
   * true if the dashboard is sorted ascending, false otherwise.
   */
  sortedAscending: ({ dashboard }: DashboardDataSources<SortingDashboard>) =>
    dashboard.sortDirection === SortDirection.ASCENDING,

  /**
   * Returns the measure name that the dashboard is sorted by,
   * or null if the dashboard is sorted by dimension value.
   */
  sortMeasure: ({ dashboard }: DashboardDataSources<SortingDashboard>) =>
    dashboard.dashboardSortType !== SortType.DIMENSION &&
    dashboard.dashboardSortType !== SortType.UNSPECIFIED
      ? dashboard.leaderboardMeasureName
      : null,

  /**
   * Returns true if the dashboard is sorted by a dimension, false otherwise.
   */
  sortedByDimensionValue: ({
    dashboard,
  }: DashboardDataSources<SortingDashboard>) =>
    dashboard.dashboardSortType === SortType.DIMENSION,
};
