import * as m from "@rilldata/web-common/paraglide/messages.js";

/**
 * Get relative time for recent dates (e.g., "2m ago", "1h ago")
 */
export function getRelativeTime(dateString: string): string {
  if (!dateString) return "";

  try {
    const date = new Date(dateString);
    const now = new Date();
    const diffInMinutes = Math.floor(
      (now.getTime() - date.getTime()) / (1000 * 60),
    );

    if (diffInMinutes < 1) return m.time_relative_now();
    if (diffInMinutes < 60)
      return m.time_relative_minutes_short({ count: diffInMinutes });

    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24)
      return m.time_relative_hours_short({ count: diffInHours });

    return "";
  } catch {
    return "";
  }
}

/**
 * Get human-readable relative time (e.g., "5 minutes ago", "2 days ago")
 */
export function timeAgo(date: Date): string {
  const now = Date.now();
  const diffMs = now - date.getTime();
  if (diffMs < 0) return m.time_just_now();

  const diffMinutes = Math.floor(diffMs / 60000);

  if (diffMinutes < 1) return m.time_just_now();

  if (diffMinutes < 60)
    return diffMinutes === 1
      ? m.time_1_minute_ago()
      : m.time_n_minutes_ago({ count: diffMinutes });

  const hours = Math.floor(diffMs / 3600000);
  if (hours < 24)
    return hours === 1
      ? m.time_1_hour_ago()
      : m.time_n_hours_ago({ count: hours });

  const days = Math.floor(diffMs / 86400000);
  if (days < 7)
    return days === 1
      ? m.time_1_day_ago()
      : m.time_n_days_ago({ count: days });

  const weeks = Math.floor(diffMs / 604800000);
  if (weeks < 5)
    return weeks === 1
      ? m.time_1_week_ago()
      : m.time_n_weeks_ago({ count: weeks });

  const months = Math.floor(diffMs / 2592000000);
  if (months < 12)
    return months === 1
      ? m.time_1_month_ago()
      : m.time_n_months_ago({ count: months });

  const years = Math.floor(diffMs / 31536000000);
  return years === 1
    ? m.time_1_year_ago()
    : m.time_n_years_ago({ count: years });
}
