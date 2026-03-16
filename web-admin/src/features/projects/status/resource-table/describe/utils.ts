import type {
  V1Schedule,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

export function formatSchedule(schedule: V1Schedule | undefined): string {
  if (!schedule) return "";
  if (schedule.cron) return schedule.cron;
  if (schedule.tickerSeconds) {
    const seconds = schedule.tickerSeconds;
    if (seconds >= 3600) {
      const hours = seconds / 3600;
      return `Every ${Number.isInteger(hours) ? hours : hours.toFixed(1)}h`;
    }
    if (seconds >= 60) {
      const minutes = seconds / 60;
      return `Every ${Number.isInteger(minutes) ? minutes : minutes.toFixed(1)}m`;
    }
    return `Every ${seconds}s`;
  }
  if (schedule.refUpdate) return "On dependency update";
  return "";
}

export function formatBytes(bytes: string | number | undefined): string {
  if (bytes === undefined || bytes === null) return "";
  const num = typeof bytes === "string" ? Number(bytes) : bytes;
  if (isNaN(num) || num === 0) return "0 B";

  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(num) / Math.log(1024));
  const value = num / Math.pow(1024, i);
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

const timeGrainLabels: Record<string, string> = {
  TIME_GRAIN_UNSPECIFIED: "Unspecified",
  TIME_GRAIN_MILLISECOND: "Millisecond",
  TIME_GRAIN_SECOND: "Second",
  TIME_GRAIN_MINUTE: "Minute",
  TIME_GRAIN_HOUR: "Hour",
  TIME_GRAIN_DAY: "Day",
  TIME_GRAIN_WEEK: "Week",
  TIME_GRAIN_MONTH: "Month",
  TIME_GRAIN_QUARTER: "Quarter",
  TIME_GRAIN_YEAR: "Year",
};

export function formatTimeGrain(grain: V1TimeGrain | undefined): string {
  if (!grain) return "";
  return timeGrainLabels[grain] ?? grain;
}

const changeModeLabels: Record<string, string> = {
  MODEL_CHANGE_MODE_UNSPECIFIED: "Unspecified",
  MODEL_CHANGE_MODE_RESET: "Reset",
  MODEL_CHANGE_MODE_MANUAL: "Manual",
  MODEL_CHANGE_MODE_PATCH: "Patch",
};

export function formatChangeMode(mode: string | undefined): string {
  if (!mode) return "";
  return changeModeLabels[mode] ?? mode;
}

const dayOfWeekLabels: Record<number, string> = {
  1: "Monday",
  2: "Tuesday",
  3: "Wednesday",
  4: "Thursday",
  5: "Friday",
  6: "Saturday",
  7: "Sunday",
};

export function formatDayOfWeek(day: number | undefined): string {
  if (day === undefined || day === null || day === 0) return "Monday";
  return dayOfWeekLabels[day] ?? String(day);
}

const monthLabels: Record<number, string> = {
  1: "January",
  2: "February",
  3: "March",
  4: "April",
  5: "May",
  6: "June",
  7: "July",
  8: "August",
  9: "September",
  10: "October",
  11: "November",
  12: "December",
};

export function formatMonthOfYear(month: number | undefined): string {
  if (month === undefined || month === null || month === 0) return "January";
  return monthLabels[month] ?? String(month);
}

/**
 * Safely formats a property value for display, handling objects that would
 * otherwise render as `[object Object]`.
 */
export function formatPropertyValue(val: unknown): string {
  if (val === undefined || val === null) return "";
  if (typeof val === "object") return JSON.stringify(val);
  return String(val);
}
