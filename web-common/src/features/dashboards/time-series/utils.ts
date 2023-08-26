import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
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
