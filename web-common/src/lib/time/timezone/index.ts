import { timeZoneNameToAbbreviationMap } from "@rilldata/web-common/lib/time/timezone/abbreviationMap";
import { DateTime } from "luxon";

export function toFormat(dt: Date, zone: string, format: string) {
  return DateTime.fromJSDate(dt).setZone(zone).toFormat(format);
}

export function removeLocalTimezoneOffset(dt: Date) {
  return new Date(dt.getTime() + dt.getTimezoneOffset() * 60000);
}

export function addZoneOffset(dt: Date, iana: string) {
  const offset = DateTime.fromJSDate(dt).setZone(iana).offset;
  return new Date(dt.getTime() + offset * 60000);
}

export function removeZoneOffset(dt: Date, iana: string) {
  const offset = DateTime.fromJSDate(dt).setZone(iana).offset;
  return new Date(dt.getTime() - offset * 60000);
}

export function getLocalIANA(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

export function getUTCIANA(): string {
  return "Etc/UTC";
}

export function getTimeZoneNameFromIANA(now: Date, iana: string): string {
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZZZZ");
}

export function getAbbreviationForIANA(now: Date, iana: string): string {
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
