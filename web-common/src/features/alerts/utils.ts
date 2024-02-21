import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import type { StateManagers } from "../dashboards/state-managers/state-managers";

export function getLabelForFieldName(ctx: StateManagers, fieldName: string) {
  const {
    selectors: {
      measures: { allMeasures },
      dimensions: { allDimensions },
    },
  } = ctx;

  const measureLabel = get(allMeasures)?.find(
    (m) => m.name === fieldName,
  )?.label;
  const dimensionLabel = get(allDimensions)?.find(
    (d) => d.name === fieldName,
  )?.label;

  return measureLabel || dimensionLabel || fieldName;
}

export function offsetTimeByGrain(end: Date, grain: V1TimeGrain) {
  return getOffset(end, TIME_GRAIN[grain].duration, TimeOffsetType.SUBTRACT);
}
