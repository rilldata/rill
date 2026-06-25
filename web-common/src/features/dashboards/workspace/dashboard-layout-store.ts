import { localStorageStore } from "@rilldata/web-common/lib/store-utils/local-storage";

// Explore view: width (px) of the timeseries charts beside the leaderboards.
export const DEFAULT_TIMESERIES_WIDTH = 580;
export const MIN_TIMESERIES_WIDTH = 440;

// Time Dimension Detail view: height (px) of the timeseries chart above the
// detail table.
export const DEFAULT_TDD_CHART_HEIGHT = 245;
export const MIN_TDD_CHART_HEIGHT = 145;

// Pivot view: width (px) of the sidebar (tags + items columns) beside the table.
// The auto width depends on whether the metrics view has tags, matching the
// legacy fixed widths (240px without tags, 400px with tags).
export const DEFAULT_PIVOT_SIDEBAR_WIDTH = 400;
export const DEFAULT_PIVOT_SIDEBAR_WIDTH_NO_TAGS = 240;
export const MIN_PIVOT_SIDEBAR_WIDTH = 240;
export const MAX_PIVOT_SIDEBAR_WIDTH = 900;

// Width bounds for the tags column shared by the dimension/measure selector
// dropdown (explore) and the pivot sidebar. In auto mode the column sizes to
// its content between MIN and CAP; dragging the divider can widen it up to
// DRAG_MAX. Pivot is additionally capped to PCT_CAP percent of the sidebar so
// the items list stays usable.
export const TAG_COLUMN = {
  explore: { MIN: 200, CAP: 340, DRAG_MAX: 360 },
  pivot: { MIN: 140, CAP: 240, DRAG_MAX: 600, PCT_CAP: 65 },
} as const;

/**
 * Width of the timeseries charts in the explore view, controlled by the
 * resizable divider between the charts and the leaderboards. Persisted so the
 * split is preserved as the user navigates back and forth.
 */
export const exploreTimeseriesWidth = localStorageStore<number>(
  "explore-timeseries-width",
  DEFAULT_TIMESERIES_WIDTH,
);

/**
 * Height of the expanded timeseries chart in the Time Dimension Detail view,
 * controlled by the resizable divider between the chart and the detail table.
 * Persisted so the split is preserved as the user navigates back and forth.
 */
export const tddChartHeight = localStorageStore<number>(
  "tdd-chart-height",
  DEFAULT_TDD_CHART_HEIGHT,
);

/**
 * Width of the pivot sidebar, controlled by the resizable divider between the
 * sidebar and the pivot table. `null` means use the tag-aware auto width; a
 * number is an explicit user-chosen width set by dragging the divider.
 * Persisted so the split is preserved as the user navigates back and forth.
 */
export const pivotSidebarWidth = localStorageStore<number | null>(
  "pivot-sidebar-width",
  null,
);

/**
 * Width of the tags column in the explore dimension/measure selector. `null`
 * means auto-fit to content (capped); a number is an explicit user-chosen width
 * set by dragging the divider. Persisted across sessions.
 */
export const exploreTagColumnWidth = localStorageStore<number | null>(
  "explore-tag-column-width",
  null,
);

/**
 * Width of the tags column in the pivot sidebar. `null` means auto-fit to
 * content (capped); a number is an explicit user-chosen width set by dragging
 * the divider. Persisted across sessions.
 */
export const pivotTagColumnWidth = localStorageStore<number | null>(
  "pivot-tag-column-width",
  null,
);
