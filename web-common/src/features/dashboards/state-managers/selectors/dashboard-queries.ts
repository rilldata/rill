import { additionalMeasures } from "../../selectors";
import type { DimensionThresholdFilter } from "../../stores/metrics-explorer-entity";

export function getMeasuresForDimensionTable(
  activeMeasureName: string,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  visibleMeasureNames: string[],
) {
  const allMeasures = new Set([
    ...visibleMeasureNames,
    // TODO: refactor activeMeasureName to activeMeasureNames
    ...additionalMeasures(activeMeasureName, dimensionThresholdFilters),
  ]);
  return [...allMeasures];
}
