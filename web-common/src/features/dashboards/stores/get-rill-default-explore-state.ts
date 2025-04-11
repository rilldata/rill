import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
import {
  getLocalIANA,
  getUTCIANA,
} from "@rilldata/web-common/lib/time/timezone";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  V1TimeGrain,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { DateTime, IANAZone, Interval } from "luxon";

// getRillDefaultExploreState to follow in a future PR. Right now our default explore has yaml config merged in.

export function getDefaultTimeRange(
  smallestTimeGrain: V1TimeGrain | undefined,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  if (!timeRangeSummary?.min || !timeRangeSummary?.max) {
    return undefined;
  }

  if (
    smallestTimeGrain &&
    smallestTimeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED
  ) {
    switch (smallestTimeGrain) {
      case V1TimeGrain.TIME_GRAIN_SECOND:
      case V1TimeGrain.TIME_GRAIN_MINUTE:
        return TimeRangePreset.LAST_SIX_HOURS;
      case V1TimeGrain.TIME_GRAIN_HOUR:
        return TimeRangePreset.LAST_24_HOURS;
      case V1TimeGrain.TIME_GRAIN_DAY:
        return TimeRangePreset.LAST_7_DAYS;
      case V1TimeGrain.TIME_GRAIN_WEEK:
        return TimeRangePreset.LAST_4_WEEKS;
      case V1TimeGrain.TIME_GRAIN_MONTH:
        return TimeRangePreset.LAST_3_MONTHS;
      case V1TimeGrain.TIME_GRAIN_YEAR:
        return "P2Y";
      default:
        return TimeRangePreset.LAST_7_DAYS;
    }
  } else {
    const dayCount = Interval.fromDateTimes(
      DateTime.fromISO(timeRangeSummary?.min),
      DateTime.fromISO(timeRangeSummary?.max),
    )
      .toDuration()
      .as("days");

    let preset: TimeRangePreset = TimeRangePreset.LAST_12_MONTHS;

    if (dayCount <= 2) {
      preset = TimeRangePreset.LAST_SIX_HOURS;
    } else if (dayCount <= 14) {
      preset = TimeRangePreset.LAST_7_DAYS;
    } else if (dayCount <= 60) {
      preset = TimeRangePreset.LAST_4_WEEKS;
    } else if (dayCount <= 180) {
      preset = TimeRangePreset.QUARTER_TO_DATE;
    }

    return preset;
  }
}

export function getDefaultTimeZone(explore: V1ExploreSpec) {
  const preference = explore.timeZones?.[0] ?? DEFAULT_TIMEZONES[0];

  if (preference === "Local") {
    return getLocalIANA();
  } else {
    try {
      const zone = new IANAZone(preference);

      if (zone.isValid) {
        return preference;
      } else {
        throw new Error("Invalid timezone");
      }
    } catch {
      return getUTCIANA();
    }
  }
}
