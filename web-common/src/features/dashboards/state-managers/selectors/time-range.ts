import type { DashboardDataSources } from "./types";
import { timeControlStateSelector } from "../../time-controls/time-control-store";

export const timeControlsState = (dashData: DashboardDataSources) =>
  timeControlStateSelector([
    dashData.metricsSpecQueryResult,
    dashData.timeRangeSummary,
    dashData.dashboard,
  ]);

export const timeRangeSelectors = {
  /**
   * Readable containing the current state of the dashboard's time controls.
   */
  timeControlsState,
};
