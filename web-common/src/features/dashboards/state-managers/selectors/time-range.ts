import {
  timeComparisonOptionsSelector,
  timeRangeSelectionsSelector,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import type { DashboardDataSources } from "./types";
import {
  selectedTimeRangeSelector,
  timeControlStateSelector,
} from "../../time-controls/time-control-store";

export const timeControlsState = (dashData: DashboardDataSources) =>
  timeControlStateSelector([
    dashData.metricsSpecQueryResult,
    dashData.timeRangeSummary,
    dashData.dashboard,
  ]);

export const isTimeControlReady = (dashData: DashboardDataSources): boolean =>
  timeControlsState(dashData).ready === true;

export const isTimeComparisonActive = (
  dashData: DashboardDataSources,
): boolean => timeControlsState(dashData).showComparison === true;

export const timeRangeSelectorState = (dashData: DashboardDataSources) =>
  timeRangeSelectionsSelector([
    dashData.metricsSpecQueryResult,
    dashData.timeRangeSummary,
    dashData.dashboard,
  ]);

export const timeComparisonOptionsState = (dashData: DashboardDataSources) =>
  timeComparisonOptionsSelector([
    dashData.metricsSpecQueryResult,
    dashData.timeRangeSummary,
    dashData.dashboard,
    selectedTimeRangeState(dashData),
  ]);

// TODO: use this in place of timeControlStore
export const selectedTimeRangeState = (dashData: DashboardDataSources) =>
  selectedTimeRangeSelector([
    dashData.metricsSpecQueryResult,
    dashData.timeRangeSummary,
    dashData.dashboard,
  ]);

export const timeRangeSelectors = {
  /**
   * Readable containing the current state of the dashboard's time controls.
   */
  timeControlsState,

  /**
   * Is the time control ready?
   */
  isTimeControlReady,

  /**
   * Is the time comparison active?
   */
  isTimeComparisonActive,

  /**
   * Selection options for the time range selector
   */
  timeRangeSelectorState,

  /**
   * Selection options for the time comparison selector
   */
  timeComparisonOptionsState,

  /**
   * Full {@link DashboardTimeControls} filled in based on selected time range.
   */
  selectedTimeRangeState,
};
