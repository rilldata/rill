import { get } from "svelte/store";
import type { StateManagers } from "../dashboards/state-managers/state-managers";

export function getLabelForFieldName(ctx: StateManagers, fieldName: string) {
  const {
    selectors: {
      measures: { allMeasures },
      dimensions: { allDimensions },
    },
  } = ctx;

  const measureLabel = get(allMeasures)?.find((m) => m.name === fieldName)
    ?.label;
  const dimensionLabel = get(allDimensions)?.find((d) => d.name === fieldName)
    ?.label;

  return measureLabel || dimensionLabel || fieldName;
}
