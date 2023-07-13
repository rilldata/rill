import { DateTime } from "luxon";

export function getLocalIANA(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

export function getAbbreviationForIANA(now: Date, iana: string): string {
  console.log(
    "getAbbreviationForIANA",
    iana,
    DateTime.fromJSDate(now).setZone(iana).zone
  );
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZZZ");
}

export function getOffsetForIANA(now: Date, iana: string): string {
  return DateTime.fromJSDate(now).setZone(iana).toFormat("ZZ");
}

export function getLabelForIANA(now: Date, iana: string): string {
  const abbreviation = getAbbreviationForIANA(now, iana);
  const offset = getOffsetForIANA(now, iana);
  return `${abbreviation} ${offset} ${iana}`;
}

export function getDateMonthYearForTimezone(date: Date, timezone: string) {
  const timeZoneDate = DateTime.fromJSDate(date).setZone(timezone);
  const day = timeZoneDate.day;
  const month = timeZoneDate.month;
  const year = timeZoneDate.year;
  return { day, month, year };
}
