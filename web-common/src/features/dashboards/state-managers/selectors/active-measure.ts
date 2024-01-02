import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { isSummableMeasure } from "../../dashboard-utils";

export const activeMeasure = (
  dashData: DashboardDataSources
): MetricsViewSpecMeasureV2 | undefined => {
  const measures = dashData.metricsSpecQueryResult.data?.measures;
  if (!measures) {
    return undefined;
  }

  const activeMeasure = measures.find(
    (measure) => measure.name === activeMeasureName(dashData)
  );
  return activeMeasure;
};

export const activeMeasureName = (dashData: DashboardDataSources): string => {
  return dashData.dashboard.leaderboardMeasureName;
};

export const selectedMeasureNames = (
  dashData: DashboardDataSources
): string[] => {
  return [...dashData.dashboard.visibleMeasureKeys];
};

export const isAnyMeasureSelected = (
  dashData: DashboardDataSources
): boolean => {
  return selectedMeasureNames(dashData).length > 0;
};

export const isValidPercentOfTotal = (
  dashData: DashboardDataSources
): boolean => {
  return activeMeasure(dashData)?.validPercentOfTotal ?? false;
};

export const activeMeasureSelectors = {
  /**
   * Gets the MetricsViewSpecMeasureV2 of the primary
   * active measure for the dashboard.
   */
  activeMeasure,
  /**
   * Gets the name of the primary active measure for the dashboard.
   */
  activeMeasureName,
  /**
   * is the currently active measure a summable measure?
   */
  isSummableMeasure: (args: DashboardDataSources) => {
    const measure = activeMeasure(args);
    return measure ? isSummableMeasure(measure) : false;
  },

  /**
   * names of the currently selected measures
   */
  selectedMeasureNames,

  /**
   * Whether any measure is currently selected
   */
  isAnyMeasureSelected,

  /**
   * Does the currently active measure have `valid_percent_of_total: true`
   * in its measure definition?
   */
  isValidPercentOfTotal,
};
