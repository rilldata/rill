import { convertTimeRangePreset } from "@rilldata/web-common/lib/time/ranges/index";
import { transformDate } from "@rilldata/web-common/lib/time/transforms";
import {
  Period,
  RelativeTimeTransformation,
  TimeOffsetType,
  TimeRange,
  TimeRangePreset,
  TimeTruncationType,
} from "@rilldata/web-common/lib/time/types";
import { Duration, parse, toString, subtract } from "duration-fns";

/**
 * Converts an ISO duration to a time range.
 * Pass in the anchor to specify when the range should be from.
 * NOTE: This should only be used for default time range. UI presets have their own settings.
 */
export function isoDurationToTimeRange(
  isoDuration: string,
  anchor: Date,
  zone = "Etc/UTC"
) {
  const startTime = transformDate(
    anchor,
    getStartTimeTransformations(isoDuration),
    zone
  );
  const endTime = transformDate(
    anchor,
    getEndTimeTransformations(isoDuration),
    zone
  );
  return {
    startTime,
    endTime,
  };
}

export const ISODurationToTimeRangePreset: Record<
  string,
  keyof typeof TimeRangePreset
> = {
  PT6H: TimeRangePreset.LAST_SIX_HOURS,
  PT24H: TimeRangePreset.LAST_24_HOURS,
  P1D: TimeRangePreset.LAST_24_HOURS,
  P7D: TimeRangePreset.LAST_7_DAYS,
  P14D: TimeRangePreset.LAST_14_DAYS,
  P4W: TimeRangePreset.LAST_4_WEEKS,
  inf: TimeRangePreset.ALL_TIME,
};
export function isoDurationToFullTimeRange(
  isoDuration: string,
  start: Date,
  end: Date,
  zone = "Etc/UTC"
): TimeRange {
  if (!isoDuration) {
    return convertTimeRangePreset(TimeRangePreset.ALL_TIME, start, end, zone);
  }
  if (isoDuration in ISODurationToTimeRangePreset) {
    return convertTimeRangePreset(
      ISODurationToTimeRangePreset[isoDuration],
      start,
      end,
      zone
    );
  }

  const { startTime, endTime } = isoDurationToTimeRange(isoDuration, end, zone);
  return {
    name: TimeRangePreset.DEFAULT,
    start: startTime,
    end: endTime,
  };
}

export function humaniseISODuration(isoDuration: string): string {
  const duration = parse(isoDuration);
  const parts = new Array<string>();
  for (let i = PeriodAndUnits.length - 1; i >= 0; i--) {
    if (!duration[PeriodAndUnits[i].unit]) continue;
    let part = duration[PeriodAndUnits[i].unit] + " " + PeriodAndUnits[i].unit;
    if (duration[PeriodAndUnits[i].unit] === 1) {
      // make the unit singular
      part = part.replace(/s$/, "");
    }
    parts.push(part);
  }

  let fullString = "";
  for (let i = 0; i < parts.length; i++) {
    fullString += parts[i];
    if (i < parts.length - 2) {
      fullString += ", ";
    } else if (i === parts.length - 2) {
      fullString += " and ";
    }
  }

  return fullString;
}

function getStartTimeTransformations(
  isoDuration: string
): Array<RelativeTimeTransformation> {
  const duration = parse(isoDuration);
  const period = getSmallestUnit(duration);
  return [
    {
      period, // this is the offset alias for the given time range alias
      truncationType: TimeTruncationType.START_OF_PERIOD,
    }, // truncation
    // then offset that by -1 of smallest period
    {
      duration: toString(subtractFromPeriod(duration, period)),
      operationType: TimeOffsetType.SUBTRACT,
    }, // operation
  ];
}

function getEndTimeTransformations(
  isoDuration: string
): Array<RelativeTimeTransformation> {
  const duration = parse(isoDuration);
  const period = getSmallestUnit(duration);
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

const PeriodAndUnits: Array<{
  period: Period;
  unit: keyof Duration;
}> = [
  {
    period: Period.MINUTE,
    unit: "minutes",
  },
  {
    period: Period.HOUR,
    unit: "hours",
  },
  {
    period: Period.DAY,
    unit: "days",
  },
  {
    period: Period.WEEK,
    unit: "weeks",
  },
  {
    period: Period.MONTH,
    unit: "months",
  },
  {
    period: Period.YEAR,
    unit: "years",
  },
];
const PeriodToUnitsMap: Partial<Record<Period, keyof Duration>> = {};
PeriodAndUnits.forEach(({ period, unit }) => (PeriodToUnitsMap[period] = unit));

function getSmallestUnit(duration: Duration) {
  for (const { period, unit } of PeriodAndUnits) {
    if (duration[unit]) {
      return period;
    }
  }

  return undefined;
}

function subtractFromPeriod(duration: Duration, period: Period) {
  if (!PeriodToUnitsMap[period]) return duration;
  return subtract(duration, { [PeriodToUnitsMap[period]]: 1 });
}
