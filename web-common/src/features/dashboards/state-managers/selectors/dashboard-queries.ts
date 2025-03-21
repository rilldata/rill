import { additionalMeasures } from "../../selectors";
import type { DimensionThresholdFilter } from "../../stores/metrics-explorer-entity";

export function getMeasuresForDimensionTable(
  activeMeasureName: string | null,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  visibleMeasureNames: string[],
) {
  const allMeasures = new Set([
    ...visibleMeasureNames,
    ...(activeMeasureName
      ? additionalMeasures(activeMeasureName, dimensionThresholdFilters)
      : []),
  ]);
  return [...allMeasures];
}
