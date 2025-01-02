import { timeComparisonOptionsSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "../../../../lib/time/types";
import {
  selectedTimeRangeSelector,
  timeControlStateSelector,
} from "../../time-controls/time-control-store";
import type { DashboardDataSources } from "./types";

export const timeControlsState = (dashData: DashboardDataSources) =>
  timeControlStateSelector([
    dashData.validMetricsView,
    dashData.validExplore,
    dashData.timeRanges,
    dashData.dashboard,
  ]);

export const isTimeControlReady = (dashData: DashboardDataSources): boolean =>
  timeControlsState(dashData).ready === true;

export const isTimeComparisonActive = (
  dashData: DashboardDataSources,
): boolean => timeControlsState(dashData).showTimeComparison === true;

export const timeComparisonOptionsState = (dashData: DashboardDataSources) =>
  timeComparisonOptionsSelector([
    dashData.validMetricsView,
    dashData.validExplore,
    dashData.timeRangeSummary,
    dashData.dashboard,
    selectedTimeRangeState(dashData),
  ]);

// TODO: use this in place of timeControlStore
export const selectedTimeRangeState = (dashData: DashboardDataSources) =>
  selectedTimeRangeSelector([
    dashData.validExplore,
    dashData.timeRangeSummary,
    dashData.dashboard,
  ]);

export const isCustomTimeRange = (dashData: DashboardDataSources): boolean =>
  dashData.dashboard?.selectedTimeRange?.name === TimeRangePreset.CUSTOM ||
  dashData.dashboard?.selectedComparisonTimeRange?.name ===
    TimeComparisonOption.CUSTOM;

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
   * Selection options for the time comparison selector
   */
  timeComparisonOptionsState,

  /**
   * Full {@link DashboardTimeControls} filled in based on selected time range.
   */
  selectedTimeRangeState,

  isCustomTimeRange,
};
