import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const allMeasures = ({
  metricsSpecQueryResult,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  const measures = metricsSpecQueryResult.data?.measures;
  return measures === undefined ? [] : measures;
};

export const visibleMeasures = ({
  metricsSpecQueryResult,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  const measures = metricsSpecQueryResult.data?.measures?.filter(
    (d) => d.name && dashboard.visibleMeasureKeys.has(d.name)
  );
  return measures === undefined ? [] : measures;
};

export const isMeasureValidPercentOfTotal = ({
  metricsSpecQueryResult,
}: DashboardDataSources): ((measureName: string) => boolean) => {
  return (measureName: string) => {
    const measures = metricsSpecQueryResult.data?.measures;
    const selectedMeasure = measures?.find((m) => m.name === measureName);
    return selectedMeasure?.validPercentOfTotal ?? false;
  };
};

export const measureSelectors = {
  /**
   * Get all measures in the dashboard.
   */
  allMeasures,

  /**
   * Gets all visible measures in the dashboard.
   */
  visibleMeasures,

  /**
   * Checks if the provided measure is a valid percent of total
   */
  isMeasureValidPercentOfTotal,
};
