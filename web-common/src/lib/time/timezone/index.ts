import { timeZoneNameToAbbreviationMap } from "@rilldata/web-common/lib/time/timezone/abbreviationMap";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
import { DateTime } from "luxon";

function getDSToffset(start: Date, end: Date) {
  const startDateTime = DateTime.fromJSDate(start);
  const endDateTime = DateTime.fromMillis(end.getTime() - 1);

  return startDateTime.offset - endDateTime.offset;
}

/**
 * Removes the local timezone offset from a date.
 *
 * Some dates in Rill are used a range rather than a point of time.
 * The runtime aggregates data and uses the start of a period to denote a range.
 * For ex. 3 Nov 2024 with a daily grain would infer to [3 Nov 2024, 4 Nov 2024)
 */
export function removeLocalTimezoneOffset(dt: Date, grainDuration = "PT0S") {
  const endTime = getOffset(dt, grainDuration, TimeOffsetType.ADD);
  const dstOffset = getDSToffset(dt, endTime);
  return new Date(dt.getTime() + (dt.getTimezoneOffset() + dstOffset) * 60000);
}

export function addZoneOffset(dt: Date, iana: string, grainDuration = "PT0S") {
  const endTime = getOffset(dt, grainDuration, TimeOffsetType.ADD);
  const dstOffset = getDSToffset(dt, endTime);

  const offset = DateTime.fromJSDate(dt).setZone(iana).offset;
  return new Date(dt.getTime() + (offset - dstOffset) * 60000);
}

export function removeZoneOffset(dt: Date, iana: string) {
  const offset = DateTime.fromJSDate(dt).setZone(iana).offset;
  return new Date(dt.getTime() - offset * 60000);
}

export function getLocalIANA(): string {
  const localIANA = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const zone = DateTime.local().setZone(localIANA);

  if (zone.isValid) return localIANA;
  else return getUTCIANA();
}

export function getUTCIANA(): string {
  return "UTC";
}

export function getTimeZoneNameFromIANA(now: Date, iana: string): string {
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZZZZ");
}

export function getAbbreviationForIANA(now: Date, iana: string) {
  const zoneName = getTimeZoneNameFromIANA(now, iana);

  if (zoneName in timeZoneNameToAbbreviationMap)
    return timeZoneNameToAbbreviationMap[zoneName];

  // fallback to the offset
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZZZ");
}

export function getOffsetForIANA(now: Date, iana: string): string {
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZ");
}

export function getHoursOffsetForIANA(now: Date, iana: string): number {
  return DateTime.fromJSDate(now).setZone(iana).offset / 60;
}

export function getLabelForIANA(now: Date, iana: string) {
  const abbreviation = getAbbreviationForIANA(now, iana);
  const offset = getOffsetForIANA(now, iana);

  return {
    abbreviation,
    offset: `GMT ${offset}`,
    iana,
  };
}

export function getDateMonthYearForTimezone(date: Date, timezone: string) {
  const timeZoneDate = DateTime.fromJSDate(date).setZone(timezone);
  const day = timeZoneDate.day;
  const month = timeZoneDate.month;
  const year = timeZoneDate.year;
  return { day, month, year };
}
