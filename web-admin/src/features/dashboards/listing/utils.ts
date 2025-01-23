import { DateTime, Duration } from "luxon";

export function timeAgo(date: Date): string {
  const now = DateTime.now();
  const then = DateTime.fromJSDate(date);
  const diff = Duration.fromMillis(now.diff(then).milliseconds);

  if (diff.as("minutes") < 1) return "just now";
  if (diff.as("hours") < 1) return `${Math.floor(diff.as("minutes"))}m ago`;
  if (diff.as("days") < 1) return `${Math.floor(diff.as("hours"))}h ago`;
  if (diff.as("weeks") < 1) return `${Math.floor(diff.as("days"))}d ago`;
  if (diff.as("months") < 1) return `${Math.floor(diff.as("weeks"))}w ago`;
  if (diff.as("years") < 1) return `${Math.floor(diff.as("months"))}M ago`;

  return `${Math.floor(diff.as("years"))}y ago`;
}
