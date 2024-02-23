import { PeriodAndUnits } from "@rilldata/web-common/lib/time/config";
import { convertTimeRangePreset } from "@rilldata/web-common/lib/time/ranges/index";
import {
  subtractFromPeriod,
  transformDate,
} from "@rilldata/web-common/lib/time/transforms";
import {
  RangePresetType,
  ReferencePoint,
  RelativeTimeTransformation,
  TimeComparisonOption,
  TimeOffsetType,
  TimeRange,
  TimeRangeMeta,
  TimeRangePreset,
  TimeTruncationType,
} from "@rilldata/web-common/lib/time/types";
import { Duration } from "luxon";

/**
 * Converts an ISO duration to a time range.
 * Pass in the anchor to specify when the range should be from.
 * NOTE: This should only be used for default time range. UI presets have their own settings.
 */
export function isoDurationToTimeRange(
  isoDuration: string,
  anchor: Date,
  zone = "UTC",
) {
  const startTime = transformDate(
    anchor,
    getStartTimeTransformations(isoDuration),
    zone,
  );
  const endTime = transformDate(
    anchor,
    getEndTimeTransformations(isoDuration),
    zone,
  );
  return {
    startTime,
    endTime,
  };
}

export const ISODurationToTimeRangePreset: Partial<
  Record<TimeRangePreset, boolean>
> = {};
for (const preset in TimeRangePreset) {
  if (preset === "DEFAULT" || preset === "CUSTOM") continue;
  ISODurationToTimeRangePreset[TimeRangePreset[preset]] = true;
}

export function isoDurationToFullTimeRange(
  isoDuration: string | undefined,
  start: Date,
  end: Date,
  zone = "UTC",
): TimeRange {
  if (!isoDuration) {
    return convertTimeRangePreset(TimeRangePreset.ALL_TIME, start, end, zone);
  }
  if (isoDuration in ISODurationToTimeRangePreset) {
    return convertTimeRangePreset(
      isoDuration as TimeRangePreset,
      start,
      end,
      zone,
    );
  }

  const { startTime, endTime } = isoDurationToTimeRange(isoDuration, end, zone);
  return {
    name: isoDuration as TimeRangePreset,
    start: startTime,
    end: endTime,
  };
}

export function humaniseISODuration(isoDuration: string): string {
  if (!isoDuration) return "";
  const duration = Duration.fromISO(isoDuration);
  let humanISO = duration.toHuman({
    listStyle: "long",
  });
  humanISO = humanISO.replace(/(\d) (\w)/g, (substring, n, c) => {
    return `${n} ${c.toUpperCase()}`;
  });
  humanISO = humanISO.replace(", and", " and");
  return humanISO;
}

export function getSmallestTimeGrain(isoDuration: string | undefined) {
  if (isoDuration === undefined) {
    return undefined;
  }

  const duration = Duration.fromISO(isoDuration);
  for (const { grain, unit } of PeriodAndUnits) {
    if (duration[unit]) {
      return grain;
    }
  }

  return undefined;
}

export function isoDurationToTimeRangeMeta(
  isoDuration: string,
  defaultComparison: TimeComparisonOption,
): TimeRangeMeta {
  return {
    label: `Last ${humaniseISODuration(isoDuration)}`,
    defaultComparison,
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: getStartTimeTransformations(isoDuration),
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: getEndTimeTransformations(isoDuration),
    },
  };
}

function getStartTimeTransformations(
  isoDuration: string,
): Array<RelativeTimeTransformation> {
  const duration = Duration.fromISO(isoDuration);
  const period = getSmallestUnit(duration);
  if (!period) return [];

  return [
    {
      period, // this is the offset alias for the given time range alias
      truncationType: TimeTruncationType.START_OF_PERIOD,
    }, // truncation
    // then offset that by -1 of smallest period
    {
      duration: subtractFromPeriod(duration, period).toISO() as string,
      operationType: TimeOffsetType.SUBTRACT,
    }, // operation
  ];
}

function getEndTimeTransformations(
  isoDuration: string,
): Array<RelativeTimeTransformation> {
  const duration = Duration.fromISO(isoDuration);
  const period = getSmallestUnit(duration);
  if (!period) return [];

  return [
    {
      duration: period,
      operationType: TimeOffsetType.ADD,
    },
    {
      period,
      truncationType: TimeTruncationType.START_OF_PERIOD,
    },
  ];
}

function getSmallestUnit(duration: Duration) {
  for (const { period, unit } of PeriodAndUnits) {
    if (duration[unit]) {
      return period;
    }
  }

  return undefined;
}
