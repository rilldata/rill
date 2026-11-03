import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime } from "luxon";

const LOCAL_ISO = "yyyy-MM-dd'T'HH:mm:ss";

// Time values are written as wall-clock times in the dashboard's time zone,
// since Excel has no notion of zones. A Date is one the chart pipeline already
// shifted (see adjustDataForTimeZone), so its local fields are that wall clock;
// an ISO string is converted here.
export function localIsoInZone(
  value: unknown,
  timeZone: string,
): string | null {
  if (value instanceof Date) {
    return DateTime.fromJSDate(value).toFormat(LOCAL_ISO);
  }
  if (typeof value === "string" && value) {
    const dt = DateTime.fromISO(value, { zone: timeZone });
    return dt.isValid ? dt.toFormat(LOCAL_ISO) : value;
  }
  return null;
}

// Whether a grain's values are plain dates (no time of day).
export function isDateGrain(grain: V1TimeGrain | undefined): boolean {
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_DAY:
    case V1TimeGrain.TIME_GRAIN_WEEK:
    case V1TimeGrain.TIME_GRAIN_MONTH:
    case V1TimeGrain.TIME_GRAIN_QUARTER:
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return true;
    default:
      return false;
  }
}

export function timeColumnType(
  grain: V1TimeGrain | undefined,
): WorkbookColumnType {
  return isDateGrain(grain)
    ? WorkbookColumnType.DATE
    : WorkbookColumnType.TIMESTAMP;
}

// The Excel number format that shows a time value at the given grain.
export function excelDateFormat(grain: V1TimeGrain | undefined): string {
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return "yyyy";
    case V1TimeGrain.TIME_GRAIN_QUARTER:
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return "mmm yyyy";
    case V1TimeGrain.TIME_GRAIN_WEEK:
    case V1TimeGrain.TIME_GRAIN_DAY:
      return "yyyy-mm-dd";
    case V1TimeGrain.TIME_GRAIN_HOUR:
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return "yyyy-mm-dd hh:mm";
    default:
      return "yyyy-mm-dd hh:mm:ss";
  }
}

// A human-readable time at the given grain, for cells that mix dimensions
// (the row header of a nested pivot) and so can't be typed as dates.
export function formatTimeLabel(
  iso: string,
  grain: V1TimeGrain | undefined,
  timeZone: string,
): string {
  const dt = DateTime.fromISO(iso, { zone: timeZone });
  if (!dt.isValid) return iso;
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return dt.toFormat("yyyy");
    case V1TimeGrain.TIME_GRAIN_QUARTER:
      return dt.toFormat("'Q'q yyyy");
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return dt.toFormat("LLL yyyy");
    case V1TimeGrain.TIME_GRAIN_WEEK:
    case V1TimeGrain.TIME_GRAIN_DAY:
      return dt.toFormat("yyyy-MM-dd");
    case V1TimeGrain.TIME_GRAIN_HOUR:
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return dt.toFormat("yyyy-MM-dd HH:mm");
    default:
      return dt.toFormat("yyyy-MM-dd HH:mm:ss");
  }
}
