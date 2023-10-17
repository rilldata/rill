import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { SelectorFnArgs } from "./types";

// (
//   dashboard: MetricsExplorerEntity,
//   metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, RpcStatus>
// )

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
