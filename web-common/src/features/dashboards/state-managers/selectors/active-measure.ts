import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { SelectorFnArgs } from "./types";
import { isSummableMeasure } from "../../dashboard-utils";

export const activeMeasure = ([
  dashboard,
  metricsSpecQueryResult,
]: SelectorFnArgs): MetricsViewSpecMeasureV2 | undefined => {
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
  activeMeasure,
  /**
   * is the currently active measure a summable measure?
   */
  isSummableMeasure: ([dashboard, metricsSpecQueryResult]: SelectorFnArgs) =>
    isSummableMeasure(activeMeasure([dashboard, metricsSpecQueryResult])),
};
