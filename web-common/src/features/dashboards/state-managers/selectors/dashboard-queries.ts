import { additionalMeasures } from "../../selectors";
import type { DimensionThresholdFilter } from "../../stores/metrics-explorer-entity";

export function getMeasuresForDimensionOrLeaderboardDisplay(
  sortByMeasureName: string | null,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  visibleMeasureNames: string[],
) {
  const allMeasures = new Set([
    ...visibleMeasureNames,
    ...(sortByMeasureName
      ? additionalMeasures(sortByMeasureName, dimensionThresholdFilters)
      : []),
  ]);
  return [...allMeasures];
}
