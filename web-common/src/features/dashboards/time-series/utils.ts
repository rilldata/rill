import { matchExpressionByName } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import {
  createAndExpression,
  removeExpressions,
} from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
import { roundToNearestTimeUnit } from "./round-to-nearest-time-unit";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
import { removeZoneOffset } from "../../../lib/time/timezone";
import { TimeOffsetType } from "../../../lib/time/types";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

/** sets extents to 0 if it makes sense; otherwise, inflates each extent component */
export function niceMeasureExtents(
  [smallest, largest]: [number, number],
  inflator: number
) {
  if (smallest === 0 && largest === 0) {
    return [0, 1];
  }
  return [
    smallest < 0 ? smallest * inflator : 0,
    largest > 0 ? largest * inflator : 0,
  ];
}

export function toComparisonKeys(d, offsetDuration: string, zone: string) {
  return Object.keys(d).reduce((acc, key) => {
    if (key === "records") {
      Object.entries(d.records).forEach(([key, value]) => {
        acc[`comparison.${key}`] = value;
      });
    } else if (`comparison.${key}` === "comparison.ts") {
      acc[`comparison.${key}`] = adjustOffsetForZone(d[key], zone);
      acc["comparison.ts_position"] = getOffset(
        acc["comparison.ts"],
        offsetDuration,
        TimeOffsetType.ADD
      );
    } else {
      acc[`comparison.${key}`] = d[key];
    }
    return acc;
  }, {});
}

export function prepareTimeSeries(
  original,
  comparison,
  timeGrainDuration: string,
  zone: string
) {
  return original?.map((originalPt, i) => {
    const comparisonPt = comparison?.[i];

    const ts = adjustOffsetForZone(originalPt.ts, zone);
    const offsetDuration = getDurationMultiple(timeGrainDuration, 0.5);
    const ts_position = getOffset(ts, offsetDuration, TimeOffsetType.ADD);
    return {
      ts,
      ts_position,
      bin: originalPt.bin,
      ...originalPt.records,
      ...toComparisonKeys(comparisonPt || {}, offsetDuration, zone),
    };
  });
}

export function getBisectedTimeFromCordinates(
  value,
  scaleStore,
  accessor,
  data,
  grainLabel
) {
  const roundedValue = roundToNearestTimeUnit(
    scaleStore.invert(value),
    grainLabel
  );
  return bisectData(roundedValue, "center", accessor, data)[accessor];
}

/**
 *  The dates in the charts are in the local timezone, this util method
 *  removes the selected timezone offset and adds the local offset
 */
export function localToTimeZoneOffset(dt: Date, zone: string) {
  const utcDate = new Date(dt.getTime() - dt.getTimezoneOffset() * 60000);
  return removeZoneOffset(utcDate, zone);
}

// Return start and end of the time range that is ordered.
export function getOrderedStartEnd(start: Date, stop: Date) {
  const startMs = start?.getTime();
  const stopMs = stop?.getTime();

  if (startMs > stopMs) {
    return { start: stop, end: start };
  } else {
    return { start, end: stop };
  }
}

export function getFilterForComparedDimension(
  dimensionName: string,
  filters: V1Expression,
  topListValues: string[]
) {
  const includedValues = topListValues?.slice(0, 250);

  let updatedFilter = removeExpressions(filters, (e) =>
    matchExpressionByName(e, dimensionName)
  );
  if (!updatedFilter) {
    updatedFilter = createAndExpression([]);
  }

  return { includedValues, updatedFilter };
}
