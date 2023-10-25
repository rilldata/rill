import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { isSummableMeasure } from "../../dashboard-utils";

export const activeMeasure = ({
  dashboard,
  metricsSpecQueryResult,
}: DashboardDataSources): MetricsViewSpecMeasureV2 | undefined => {
  const measures = metricsSpecQueryResult.data?.measures;
  if (!measures) {
    return undefined;
  }

  const activeMeasure = measures.find(
    (measure) => measure.name === dashboard.leaderboardMeasureName
  );
  return activeMeasure;
};

export const activeMeasureSelectors = {
  /**
   * Gets the active measure for the dashboard.
   */
  activeMeasure,
  /**
   * is the currently active measure a summable measure?
   */
  isSummableMeasure: (args: DashboardDataSources) => {
    const measure = activeMeasure(args);
    return measure ? isSummableMeasure(measure) : false;
  },
};
