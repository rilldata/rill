import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { V1Conversation } from "../../../../runtime-client";

/**
 * Group conversations by date categories (Today, Yesterday, etc.)
 */
export function groupConversationsByDate(conversations: V1Conversation[]): {
  [key: string]: V1Conversation[];
} {
  const groups: { [key: string]: V1Conversation[] } = {};
  const now = new Date();

  // Get today's date at midnight in local timezone
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

  conversations.forEach((conv) => {
    const dateStr = conv.updatedOn || conv.createdOn || "";
    if (!dateStr) return;

    const date = new Date(dateStr);
    // Get conversation date at midnight in local timezone
    const convDate = new Date(
      date.getFullYear(),
      date.getMonth(),
      date.getDate(),
    );

    // Calculate difference in calendar days
    const diffInDays = Math.floor(
      (today.getTime() - convDate.getTime()) / (1000 * 60 * 60 * 24),
    );

    let groupKey: string;
    if (diffInDays === 0) {
      groupKey = "Today";
    } else if (diffInDays === 1) {
      groupKey = "Yesterday";
    } else if (diffInDays < 7) {
      groupKey = `${diffInDays}d ago`;
    } else {
      groupKey = "Older";
    }

    if (!groups[groupKey]) {
      groups[groupKey] = [];
    }
    groups[groupKey].push(conv);
  });

  // Sort conversations within each group by date (newest first)
  Object.keys(groups).forEach((key) => {
    groups[key].sort((a, b) => {
      const dateA = new Date(a.updatedOn || a.createdOn || "");
      const dateB = new Date(b.updatedOn || b.createdOn || "");
      return dateB.getTime() - dateA.getTime();
    });
  });

  return groups;
}

/**
 * Standard group order for consistent conversation list display
 */
export const GROUP_ORDER = [
  "Today",
  "Yesterday",
  "2d ago",
  "3d ago",
  "4d ago",
  "5d ago",
  "6d ago",
  "Older",
] as const;

/**
 * Translate a group key into a localised display label.
 * Must be called at render time (not module scope) so Paraglide can react to language changes.
 */
export function getGroupLabel(key: string): string {
  switch (key) {
    case "Today":
      return m.chat_group_today();
    case "Yesterday":
      return m.chat_group_yesterday();
    case "Older":
      return m.chat_group_older();
    default: {
      const match = key.match(/^(\d+)d ago$/);
      if (match) {
        return m.chat_group_days_ago({ days: match[1] });
      }
      return key;
    }
  }
}
