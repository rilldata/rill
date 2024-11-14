import { getIndependentMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { additionalMeasures } from "../../selectors";
import type { DimensionThresholdFilter } from "../../stores/metrics-explorer-entity";

export function getMeasuresForDimensionTable(
  activeMeasureName: string,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  metricsView: V1MetricsViewSpec | undefined,
  visibleMeasureNames: string[],
) {
  const allMeasures = new Set([
    ...visibleMeasureNames,
    ...additionalMeasures(activeMeasureName, dimensionThresholdFilters),
  ]);
  return getIndependentMeasures(metricsView ?? {}, [...allMeasures]);
}
