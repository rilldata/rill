import { DateTime } from "luxon";

export function formatRunDate(lastRun: string, timeZone: string) {
  const luxonDate = DateTime.fromISO(lastRun, { zone: timeZone || "UTC" });

  return luxonDate.toFormat("MMM dd, yyyy, h:mm a ZZZZ");
}
