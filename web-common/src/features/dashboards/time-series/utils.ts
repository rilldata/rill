import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
import { roundToNearestTimeUnit } from "./round-to-nearest-time-unit";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
import { removeZoneOffset } from "../../../lib/time/timezone";
import { TimeOffsetType } from "../../../lib/time/types";
import {DateTime} from "luxon";

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
  zone: string,
  start: string,
  end: string,
  compStart?: string,
  compEnd?: string
): any[] {
  let i = 0;
  let j = 0;
  let k = 0;
  let dtStart = DateTime.fromISO(start, {zone});
  dtStart = dtStart.startOf("hour")
  let dtEnd = DateTime.fromISO(end, {zone}).startOf("hour");
  let dtCompStart = DateTime.fromISO(compStart, {zone}).startOf("hour"); 
  let dtCompEnd = DateTime.fromISO(compEnd, {zone}).startOf("hour"); 

  let result = [];

  const offsetDuration = getDurationMultiple(timeGrainDuration, 0.5);
  while (dtStart < dtEnd || dtCompStart < dtCompEnd) {
    let ts = adjustOffsetForZone(dtStart.toISO(), zone);
    let ts_position = getOffset(ts, offsetDuration, TimeOffsetType.ADD);
    result.push({
      ts,
      ts_position,
    });

    if (i < original.length && dtStart.equals(DateTime.fromISO(original[i].ts, {zone}))) {
      result[j] = {
        ...result[j],
        ...original[i].records,
      };
      i++;
    } 
    if (comparison) {
      if (k < comparison.length && dtCompStart.equals(DateTime.fromISO(comparison[k].ts, {zone}))) {
        result[j] = {
          ...result[j],
          ...toComparisonKeys(comparison[k], offsetDuration, zone),
        };
        k++;
      } else {
        result[j] = {
            ...result[j],
            ...toComparisonKeys({
              ts: dtCompStart.toISO()
            }, offsetDuration, zone),
          };
      }
    }
    dtStart = dtStart.plus({hours: 1});
    dtCompStart = dtCompStart.plus({hours: 1});
    j++;
  }

  return result;
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
