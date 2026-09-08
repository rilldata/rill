import { additionalMeasures } from "../../selectors";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function getMeasuresForDimensionOrLeaderboardDisplay(
  sortByMeasureName: string | null,
  expr: V1Expression | undefined,
  visibleMeasureNames: string[],
) {
  const allMeasures = new Set([
    ...visibleMeasureNames,
    ...(sortByMeasureName ? additionalMeasures(sortByMeasureName, expr) : []),
  ]);
  return [...allMeasures];
}
