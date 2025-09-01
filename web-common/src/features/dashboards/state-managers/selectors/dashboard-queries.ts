import { additionalMeasures } from "../../selectors";
import type { DimensionThresholdFilter } from "web-common/src/features/dashboards/stores/explore-state";

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
