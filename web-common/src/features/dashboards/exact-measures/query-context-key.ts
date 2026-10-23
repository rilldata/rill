import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

/**
 * Serializes the parts of the explore state that change which queries the
 * dashboard issues: filters, time controls, the active view and its
 * configuration, and sorting. The exact-measure selection is tied to this key:
 * when the key changes, the selection resets so new queries are approximate
 * again and exact values are only fetched after an explicit click.
 *
 * Presentation-only state is excluded on purpose: showing, hiding or reordering
 * measures and dimensions, context columns, chart type, pinned TDD values,
 * pivot pagination and pivot row expansion keep the selection.
 */
export function getQueryContextKey(state: ExploreState): string {
  return JSON.stringify({
    where: state.whereFilter,
    thresholds: state.dimensionThresholdFilters,
    excludeMode: [...(state.dimensionFilterExcludeMode?.entries() ?? [])],
    timeRange: state.selectedTimeRange,
    comparisonTimeRange: state.selectedComparisonTimeRange,
    showTimeComparison: state.showTimeComparison,
    timezone: state.selectedTimezone,
    timeDimension: state.selectedTimeDimension,
    scrub: state.selectedScrubRange,
    page: state.activePage,
    dimension: state.selectedDimensionName,
    comparisonDimension: state.selectedComparisonDimension,
    tddMeasure: state.tdd?.expandedMeasureName,
    pivot: state.pivot
      ? {
          rows: state.pivot.rows,
          columns: state.pivot.columns,
          sorting: state.pivot.sorting,
          enableComparison: state.pivot.enableComparison,
          tableMode: state.pivot.tableMode,
          rowLimit: state.pivot.rowLimit,
          outermostRowLimit: state.pivot.outermostRowLimit,
        }
      : undefined,
    sortBy: state.leaderboardSortByMeasureName,
    sortDirection: state.sortDirection,
    sortType: state.dashboardSortType,
  });
}
