import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
import { roundToNearestTimeUnit } from "./round-to-nearest-time-unit";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
import { removeZoneOffset } from "../../../lib/time/timezone";
import { TimeOffsetType } from "../../../lib/time/types";
import type { V1MetricsViewFilter } from "@rilldata/web-common/runtime-client";

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
    result.push({
      ts,
      ts_position,
    });

    if (
      i < original.length &&
      dtStart.equals(DateTime.fromISO(original[i].ts, { zone }))
    ) {
      result[j] = {
        ...result[j],
        ...original[i].records,
      };
      i++;
    }
    if (comparison) {
      if (
        k < comparison.length &&
        dtCompStart.equals(DateTime.fromISO(comparison[k].ts, { zone }))
      ) {
        result[j] = {
          ...result[j],
          ...toComparisonKeys(comparison[k], offsetDuration, zone),
        };
        k++;
      } else {
        result[j] = {
          ...result[j],
          ...toComparisonKeys(
            {
              ts: dtCompStart.toISO(),
            },
            offsetDuration,
            zone
          ),
        };
      }
    }

    switch (dtu) {
      case "year":
        dtStart = dtStart.plus({ years: 1 });
        dtCompStart = dtCompStart.plus({ years: 1 });
        break;
      case "quarter":
        dtStart = dtStart.plus({ quarters: 1 });
        dtCompStart = dtCompStart.plus({ quarters: 1 });
        break;
      case "month":
        dtStart = dtStart.plus({ months: 1 });
        dtCompStart = dtCompStart.plus({ months: 1 });
        break;
      case "week":
        dtStart = dtStart.plus({ weeks: 1 });
        dtCompStart = dtCompStart.plus({ weeks: 1 });
        break;
      case "day":
        dtStart = dtStart.plus({ days: 1 });
        dtCompStart = dtCompStart.plus({ days: 1 });
        break;
      case "hour":
        dtStart = dtStart.plus({ hours: 1 });
        dtCompStart = dtCompStart.plus({ hours: 1 });
        break;
      case "minute":
        dtStart = dtStart.plus({ minutes: 1 });
        dtCompStart = dtCompStart.plus({ minutes: 1 });
        break;
      case "second":
        dtStart = dtStart.plus({ seconds: 1 });
        dtCompStart = dtCompStart.plus({ seconds: 1 });
        break;
      case "millisecond":
        dtStart = dtStart.plus({ milliseconds: 1 });
        dtCompStart = dtCompStart.plus({ milliseconds: 1 });
        break;
    }
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

export function getFilterForComparedDimension(
  dimensionName: string,
  filters: V1MetricsViewFilter,
  topListValues: string[],
  surface
) {
  // Check if we have an excluded filter for the dimension
  const excludedFilter = filters.exclude.find((d) => d.name === dimensionName);

  let excludedValues = [];
  if (excludedFilter) {
    excludedValues = excludedFilter.in;
  }
  let includedValues;

  if (surface === "table") {
    // TODO : make this configurable
    includedValues = topListValues?.slice(0, 250);
  } else {
    // Remove excluded values from top list
    includedValues = topListValues
      ?.filter((d) => !excludedValues.includes(d))
      ?.slice(0, 3);
  }

  // Add dimension to filter
  const updatedFilter = {
    ...filters,
    include: [
      ...filters.include,
      {
        name: dimensionName,
        in: [],
      },
    ],
  };

  return { includedValues, updatedFilter };
}
