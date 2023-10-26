import { DateTime, Duration } from "luxon";

export function timeAgo(date: Date): string {
  const now = DateTime.now();
  const then = DateTime.fromJSDate(date);
  const diff = Duration.fromMillis(now.diff(then).milliseconds);

  if (diff.as("minutes") < 1) return "just now";
  if (diff.as("hours") < 1)
    return `${Math.floor(diff.as("minutes"))} min${
      Math.floor(diff.as("minutes")) > 1 ? "s" : ""
    } ago`;
  if (diff.as("days") < 1)
    return `${Math.floor(diff.as("hours"))} hour${
      Math.floor(diff.as("hours")) > 1 ? "s" : ""
    } ago`;
  if (diff.as("weeks") < 1)
    return `${Math.floor(diff.as("days"))} day${
      Math.floor(diff.as("days")) > 1 ? "s" : ""
    } ago`;
  if (diff.as("months") < 1)
    return `${Math.floor(diff.as("weeks"))} week${
      Math.floor(diff.as("weeks")) > 1 ? "s" : ""
    } ago`;
  if (diff.as("years") < 1)
    return `${Math.floor(diff.as("months"))} month${
      Math.floor(diff.as("months")) > 1 ? "s" : ""
    } ago`;

  return `${Math.floor(diff.as("years"))} year${
    Math.floor(diff.as("years")) > 1 ? "s" : ""
  } ago`;
}
