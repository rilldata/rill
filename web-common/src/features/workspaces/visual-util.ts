import {
  LATEST_WINDOW_TIME_RANGES,
  PERIOD_TO_DATE_RANGES,
  PREVIOUS_COMPLETE_DATE_RANGES,
} from "@rilldata/web-common/lib/time/config";

export function isString(value: unknown): value is string {
  return typeof value === "string";
}

export function stringGuard(value: unknown | undefined): string {
  return value && typeof value === "string" ? value : "";
}

export function numberGuard(value: unknown | undefined): number | undefined {
  return value && typeof value === "number" ? value : undefined;
}

export const DEFAULT_RANGES = [
  ...Object.keys(LATEST_WINDOW_TIME_RANGES),
  ...Object.keys(PERIOD_TO_DATE_RANGES),
  ...Object.keys(PREVIOUS_COMPLETE_DATE_RANGES),
];
