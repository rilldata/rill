import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const activeMeasure = (
  dashData: DashboardDataSources,
): MetricsViewSpecMeasure | undefined => {
  if (!dashData.validMetricsView?.measures) {
    return undefined;
  }

  const activeMeasure = dashData.validMetricsView.measures.find(
    (measure) => measure.name === activeMeasureName(dashData),
  );
  return activeMeasure;
};

export const activeMeasureName = (dashData: DashboardDataSources): string => {
  return dashData.dashboard.leaderboardSortByMeasureName;
};

export const selectedMeasureNames = (
  dashData: DashboardDataSources,
): string[] => {
  return dashData.dashboard.visibleMeasures;
};

export const isValidPercentOfTotal = (
  dashData: DashboardDataSources,
): boolean => {
  return activeMeasure(dashData)?.validPercentOfTotal ?? false;
};

export const activeMeasureSelectors = {
  /**
   * Gets the MetricsViewSpecMeasure of the primary
   * active measure for the dashboard.
   */
  activeMeasure,
  /**
   * Gets the name of the primary active measure for the dashboard.
   */
  activeMeasureName,

  /**
   * names of the currently selected measures
   */
  selectedMeasureNames,

  /**
   * Does the currently active measure have `valid_percent_of_total: true`
   * in its measure definition?
   */
  isValidPercentOfTotal,
};
