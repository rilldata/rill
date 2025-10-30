import { getSmallestUnitInDateTime } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import {
  type V1Expression,
  V1Operation,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { DateTime } from "luxon";

export function parseTimeRangeFromFilters(
  filter: V1Expression,
  timeDimension: string,
  timezone: string,
  timeRangeSummary: V1TimeRangeSummary,
) {
  let start: DateTime | undefined = undefined;
  let end: DateTime | undefined = undefined;

  const maybeSetTime = (expr: V1Expression, isStart: boolean) => {
    if (!expr.cond?.exprs) return false;
    const ident = expr.cond.exprs[0]?.ident;
    if (ident !== timeDimension) return false;

    const val = expr.cond.exprs?.[1]?.val;
    const valDt = DateTime.fromISO(val as string).setZone(timezone);
    if (!valDt.isValid) return false;

    if (isStart && !start) {
      start = valDt;
      return true;
    }

    if (!isStart && !end) {
      end = valDt;
      return true;
    }

    return false;
  };

  const list = [filter];

  while (list.length > 0) {
    const f = list.shift();
    if (!f?.cond?.op) continue;

    switch (f.cond.op) {
      case V1Operation.OPERATION_OR:
        break;

      case V1Operation.OPERATION_AND:
        list.push(...(f.cond.exprs ?? []));
        break;

      case V1Operation.OPERATION_EQ:
        if (maybeSetTime(f, true) && start) {
          end = offsetMinUnit(start);
        }
        break;

      case V1Operation.OPERATION_GT:
        if (maybeSetTime(f, true) && start) {
          start = offsetMinUnit(start);
        }
        break;

      case V1Operation.OPERATION_GTE:
        maybeSetTime(f, true);
        break;

      case V1Operation.OPERATION_LT:
        maybeSetTime(f, false);
        break;

      case V1Operation.OPERATION_LTE:
        if (maybeSetTime(f, false) && end) {
          end = offsetMinUnit(end);
        }
        break;

      default:
        break;
    }
  }

  if (!end && start && timeRangeSummary.max) {
    end = DateTime.fromISO(timeRangeSummary.max);
  }
  if (!start && end && timeRangeSummary.min) {
    start = DateTime.fromISO(timeRangeSummary.min);
  }

  if (!end || !start) return undefined;

  return {
    name: TimeRangePreset.CUSTOM,
    start: start.toJSDate(),
    end: end.toJSDate(),
  };
}

function offsetMinUnit(time: DateTime): DateTime {
  const smallestUnit = getSmallestUnitInDateTime(time);
  if (!smallestUnit) return time;
  return time.minus({ [smallestUnit]: -1 });
}
