import { DateTime } from "luxon";
import { TimeRangePreset } from "../types";

export function getRange(params: {
  queryStart: string | null;
  queryEnd: string | null;
  min: string;
  max: string;
  period?: TimeRangePreset;
  zone?: string;
}): { start: DateTime; end: DateTime } {
  const zone = params.zone ?? "UTC";
  const min = DateTime.fromISO(params.min, { zone });
  const max = DateTime.fromISO(params.max, { zone });
  const queryStart = params.queryStart
    ? DateTime.fromISO(params.queryStart)
    : null;
  const queryEnd = params.queryEnd ? DateTime.fromISO(params.queryEnd) : null;
  const period = params.period;

  if (queryStart && queryEnd) {
    return {
      start: queryStart,
      end: queryEnd,
    };
  }

  if (queryStart) {
    return {
      start: queryStart,
      end: max,
    };
  }

  if (queryEnd) {
    return {
      start: min,
      end: queryEnd,
    };
  }

  switch (period) {
    case TimeRangePreset.ALL_TIME:
    case undefined:
      return {
        start: min,
        end: max,
      };
    case TimeRangePreset.LAST_SIX_HOURS:
      return {
        start: max.minus({ hours: 6 }),
        end: max,
      };
    case TimeRangePreset.LAST_24_HOURS:
      return {
        start: max.minus({ days: 1 }),
        end: max,
      };
    case TimeRangePreset.LAST_7_DAYS:
      return {
        start: max.minus({ days: 7 }),
        end: max,
      };
    case TimeRangePreset.LAST_14_DAYS:
      return {
        start: max.minus({ days: 14 }),
        end: max,
      };
    case TimeRangePreset.LAST_4_WEEKS:
      return {
        start: max.minus({ weeks: 4 }),
        end: max,
      };
    case TimeRangePreset.LAST_12_MONTHS:
      return {
        start: max.minus({ months: 12 }),
        end: max,
      };
    case TimeRangePreset.TODAY:
      return {
        start: DateTime.now().setZone(zone).startOf("day"),
        end: DateTime.now().setZone(zone).endOf("day"),
      };
    case TimeRangePreset.WEEK_TO_DATE:
      return {
        start: DateTime.now().setZone(zone).startOf("week"),
        end: DateTime.now().setZone(zone).endOf("day"),
      };
    case TimeRangePreset.MONTH_TO_DATE:
      return {
        start: DateTime.now().setZone(zone).startOf("month"),
        end: DateTime.now().setZone(zone).endOf("day"),
      };
    case TimeRangePreset.QUARTER_TO_DATE:
      return {
        start: DateTime.now().setZone(zone).startOf("quarter"),
        end: DateTime.now().setZone(zone).endOf("day"),
      };
    case TimeRangePreset.YEAR_TO_DATE:
      return {
        start: DateTime.now().setZone(zone).startOf("year"),
        end: DateTime.now().setZone(zone).endOf("day"),
      };
    case TimeRangePreset.YESTERDAY_COMPLETE:
      return {
        start: DateTime.now().setZone(zone).minus({ days: 1 }).startOf("day"),
        end: DateTime.now().setZone(zone).minus({ days: 1 }).endOf("day"),
      };
    case TimeRangePreset.PREVIOUS_WEEK_COMPLETE:
      return {
        start: DateTime.now().setZone(zone).minus({ weeks: 1 }).startOf("week"),
        end: DateTime.now().setZone(zone).minus({ weeks: 1 }).endOf("week"),
      };
    case TimeRangePreset.PREVIOUS_MONTH_COMPLETE:
      return {
        start: DateTime.now()
          .setZone(zone)
          .minus({ months: 1 })
          .startOf("month"),
        end: DateTime.now().setZone(zone).minus({ months: 1 }).endOf("month"),
      };
    case TimeRangePreset.PREVIOUS_QUARTER_COMPLETE:
      return {
        start: DateTime.now()
          .setZone(zone)
          .minus({ quarters: 1 })
          .startOf("quarter"),
        end: DateTime.now()
          .setZone(zone)
          .minus({ quarters: 1 })
          .endOf("quarter"),
      };

    default:
      return {
        start: min,
        end: max,
      };
  }
}
