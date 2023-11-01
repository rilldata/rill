import type { DashboardDataSources } from "./types";
import { timeControlStateSelector } from "../../time-controls/time-control-store";

export const timeControlsState = (dashData: DashboardDataSources) =>
  timeControlStateSelector([
    dashData.metricsSpecQueryResult,
    dashData.timeRangeSummary,
    dashData.dashboard,
  ]);

export const isTimeControlReady = (dashData: DashboardDataSources): boolean =>
  timeControlsState(dashData).ready === true;

export const isTimeComparisonActive = (
  dashData: DashboardDataSources
): boolean => timeControlsState(dashData).showComparison === true;

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
};
