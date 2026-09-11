import { appendEphemeralSpecMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const activeMeasure = (
  dashData: DashboardDataSources,
): MetricsViewSpecMeasure | undefined => {
  if (!dashData.validMetricsView?.measures) {
    return undefined;
  }

  // Ephemeral measures can be the sort measure too; they have no spec entry,
  // so synthesize one for formatters and tooltips.
  const name = activeMeasureName(dashData);
  return appendEphemeralSpecMeasures(
    dashData.validMetricsView.measures,
    dashData.dashboard.ephemeralMeasures,
  ).find((measure) => measure.name === name);
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
