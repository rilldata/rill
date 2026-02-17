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

    if (diffInMinutes < 1) return "now";
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;

    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) return `${diffInHours}h ago`;

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
  const diffMinutes = Math.round(diffMs / 60000);

  if (diffMinutes < 1) return "Just now";

  if (diffMinutes < 60)
    return `${diffMinutes} ${diffMinutes === 1 ? "minute" : "minutes"} ago`;

  const hours = Math.round(diffMs / 3600000);
  if (hours < 24) return `${hours} ${hours === 1 ? "hour" : "hours"} ago`;

  const days = Math.round(diffMs / 86400000);
  if (days < 7) return `${days} ${days === 1 ? "day" : "days"} ago`;

  const weeks = Math.round(diffMs / 604800000);
  if (weeks < 5) return `${weeks} ${weeks === 1 ? "week" : "weeks"} ago`;

  const months = Math.round(diffMs / 2592000000);
  if (months < 12) return `${months} ${months === 1 ? "month" : "months"} ago`;

  const years = Math.round(diffMs / 31536000000);
  return `${years} ${years === 1 ? "year" : "years"} ago`;
}
