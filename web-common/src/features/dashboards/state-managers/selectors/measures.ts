import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const visibleMeasures = ({
  metricsSpecQueryResult,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  const measures = metricsSpecQueryResult.data?.measures?.filter(
    (d) => d.name && dashboard.visibleMeasureKeys.has(d.name)
  );
  return measures === undefined ? [] : measures;
};

export const measureSelectors = {
  /**
   * Gets all visible measures in the dashboard.
   */
  visibleMeasures,
};
