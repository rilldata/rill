/**
 * Util methods for handling vega signals
 */

import type { TimeRange } from "@rilldata/web-common/lib/time/types";

export function resolveSignalField(value: unknown, field: string) {
  if (typeof value === "object" && value !== null) {
    return Array.isArray(value[field]) ? value[field][0] : undefined;
  }
  return undefined;
}

export function resolveSignalTimeField(value: unknown) {
  /**
   * Time fields end with `_ts`
   * We iterate over the keys of the object and return the first key that ends with `_ts`
   * */
  if (typeof value === "object" && value !== null) {
    for (const key in value) {
      if (key.endsWith("_ts")) {
        const ts = resolveSignalField(value, key);

        if (ts !== undefined) {
          return new Date(ts);
        }
      }
    }
  }
  return undefined;
}

export function resolveSignalIntervalField(
  value: unknown,
): TimeRange | undefined {
  /**
   * Time range fields can be either 'ts' or end with '_ts'
   * We check for both cases and return a TimeRange if a valid array of two timestamps is found.
   */
  if (typeof value === "object" && value !== null) {
    const checkAndCreateTimeRange = (arr: unknown): TimeRange | undefined => {
      if (Array.isArray(arr) && arr.length === 2) {
        const [start, end] = arr;
        return {
          start: new Date(start),
          end: new Date(end),
        };
      }
      return undefined;
    };

    // Check for 'ts' key first
    if ("ts" in value) {
      return checkAndCreateTimeRange(value["ts"]);
    }

    // If 'ts' is not found, check for keys ending with '_ts'
    for (const key in value) {
      if (key.endsWith("_ts")) {
        const timeRange = checkAndCreateTimeRange(value[key]);
        if (timeRange) return timeRange;
      }
    }
  }
  return undefined;
}
