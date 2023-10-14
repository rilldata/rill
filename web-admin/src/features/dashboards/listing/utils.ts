export function timeAgo(timestamp: number): string {
  const now: number = new Date().getTime();
  const diff: number = now - timestamp;
  const minute: number = 60 * 1000;
  const hour: number = 60 * minute;
  const day: number = 24 * hour;
  const week: number = 7 * day;
  const month: number = 30 * day;
  const year: number = 365 * day;

  if (diff < minute) return "just now";
  if (diff < hour)
    return `${Math.floor(diff / minute)} min${
      Math.floor(diff / minute) > 1 ? "s" : ""
    } ago`;
  if (diff < day)
    return `${Math.floor(diff / hour)} hour${
      Math.floor(diff / hour) > 1 ? "s" : ""
    } ago`;
  if (diff < week)
    return `${Math.floor(diff / day)} day${
      Math.floor(diff / day) > 1 ? "s" : ""
    } ago`;
  if (diff < month)
    return `${Math.floor(diff / week)} week${
      Math.floor(diff / week) > 1 ? "s" : ""
    } ago`;
  if (diff < year)
    return `${Math.floor(diff / month)} month${
      Math.floor(diff / month) > 1 ? "s" : ""
    } ago`;
  return `${Math.floor(diff / year)} year${
    Math.floor(diff / year) > 1 ? "s" : ""
  } ago`;
}
