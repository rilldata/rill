import { DateTime, Duration } from "luxon";

function formatUnit(value: number, unit: string): string {
  return `${value} ${value === 1 ? unit : unit + "s"} ago`;
}

export function timeAgo(date: Date): string {
  const now = DateTime.now();
  const then = DateTime.fromJSDate(date);
  const diff = Duration.fromMillis(now.diff(then).milliseconds);

  if (diff.as("minutes") < 1) return "just now";

  const minutes = Math.round(diff.as("minutes"));
  if (diff.as("hours") < 1) return formatUnit(minutes, "minute");

  const hours = Math.round(diff.as("hours"));
  if (diff.as("days") < 1) return formatUnit(hours, "hour");

  const days = Math.round(diff.as("days"));
  if (diff.as("weeks") < 1) return formatUnit(days, "day");

  const weeks = Math.round(diff.as("weeks"));
  if (diff.as("months") < 1) return formatUnit(weeks, "week");

  const months = Math.round(diff.as("months"));
  if (diff.as("years") < 1) return formatUnit(months, "month");

  const years = Math.round(diff.as("years"));
  return formatUnit(years, "year");
}
