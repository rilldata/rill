import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
import {
  createAndExpression,
  filterExpressions,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import type {
  V1Expression,
  V1TimeSeriesValue,
} from "@rilldata/web-common/runtime-client";
import { removeZoneOffset } from "../../../lib/time/timezone";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
import { TimeOffsetType } from "../../../lib/time/types";
import { roundToNearestTimeUnit } from "./round-to-nearest-time-unit";
import type { GraphicScale } from "@rilldata/web-common/components/data-graphic/state/types";
import type { TimeSeriesDatum } from "./timeseries-data-store";
import type { DateTimeUnit } from "luxon";

/** sets extents to 0 if it makes sense; otherwise, inflates each extent component */
export function niceMeasureExtents(
  [smallest, largest]: [number, number],
  inflator: number,
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
        TimeOffsetType.ADD,
      );
    } else {
      acc[`comparison.${key}`] = d[key];
    }
    return acc;
  }, {});
}

export function prepareTimeSeries(
  original: V1TimeSeriesValue[],
  comparison: V1TimeSeriesValue[] | undefined,
  timeGrainDuration: string,
  zone: string,
) {
  return original?.map((originalPt, i) => {
    const comparisonPt = comparison?.[i];

    const emptyPt = {
      ts: undefined,
      ts_position: undefined,
      bin: undefined,
      ...originalPt.records,
    };

    if (!originalPt?.ts) {
      return emptyPt;
    }
    const ts = adjustOffsetForZone(originalPt.ts, zone);
    if (!ts || typeof ts === "string") {
      return emptyPt;
    }
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
  value: number,
  scaleStore: GraphicScale,
  accessor: string,
  data: TimeSeriesDatum[],
  grainLabel: DateTimeUnit,
): Date | null {
  const roundedValue = roundToNearestTimeUnit(
    new Date(scaleStore.invert(value)),
    grainLabel,
  );
  const { entry: bisector } = bisectData(
    roundedValue,
    "center",
    accessor,
    data,
  );
  if (!bisector || typeof bisector === "number") return null;
  const bisected = bisector[accessor];
  if (!bisected) return null;

  return new Date(bisected);
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
  topListValues: string[],
) {
  const includedValues = topListValues?.slice(0, 250);

  let updatedFilter = filterExpressions(
    filters,
    (e) => !matchExpressionByName(e, dimensionName),
  );
  if (!updatedFilter) {
    updatedFilter = createAndExpression([]);
  }

  return { includedValues, updatedFilter };
}
