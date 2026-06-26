import { localStorageStore } from "@rilldata/web-common/lib/store-utils/local-storage";

// Explore view: width (px) of the timeseries charts beside the leaderboards.
export const DEFAULT_TIMESERIES_WIDTH = 580;
export const MIN_TIMESERIES_WIDTH = 440;

// Time Dimension Detail view: height (px) of the timeseries chart above the
// detail table.
export const DEFAULT_TDD_CHART_HEIGHT = 245;
export const MIN_TDD_CHART_HEIGHT = 145;

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
