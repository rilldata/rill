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
    (d) => d.name && dashboard.visibleMeasureKeys.has(d.name),
  );
  return measures === undefined ? [] : measures;
};

export const getMeasureByName = (
  dashData: DashboardDataSources,
): ((name: string) => MetricsViewSpecMeasureV2 | undefined) => {
  return (name: string) => {
    return allMeasures(dashData)?.find((measure) => measure.name === name);
  };
};

export const measureLabel = ({
  metricsSpecQueryResult,
}: DashboardDataSources): ((m: string) => string) => {
  return (measureName) => {
    const measure = metricsSpecQueryResult.data?.measures?.find(
      (d) => d.name === measureName,
    );
    return measure?.label ?? measureName;
  };
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
   * Returns a function that can be used to get a MetricsViewSpecMeasureV2
   * by name; this fn returns undefined if the dashboard has no measure with that name.
   */
  getMeasureByName,

  /**
   * Gets all visible measures in the dashboard.
   */
  visibleMeasures,
  /**
   * Get label for a measure by name
   */
  measureLabel,
  /**
   * Checks if the provided measure is a valid percent of total
   */
  isMeasureValidPercentOfTotal,
};
