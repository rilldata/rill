import { DateTime } from "luxon";

/**
 * Checks if a value is a valid Luxon DateTime or can be parsed into one.
 * Accepts ISO strings, JS Date, or Luxon DateTime.
 */
export function isValidDateTime(value: unknown): boolean {
  if (!value) return false;
  if (value instanceof DateTime) return value.isValid;
  if (value instanceof Date) return DateTime.fromJSDate(value).isValid;
  if (typeof value === "string") return DateTime.fromISO(value).isValid;
  return false;
}
