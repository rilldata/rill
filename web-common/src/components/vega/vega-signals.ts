/**
 * Util methods for handling vega signals
 */

export function resolveSignalField(value: unknown, field: string) {
  if (typeof value === "object" && value !== null) {
    return Array.isArray(value[field]) ? value[field][0] : undefined;
  }
  return undefined;
}

export function resolveSignalTimeField(value: unknown, temporalField?: string) {
  if (typeof value !== "object" || value === null) {
    return undefined;
  }

  // When a temporal field name is provided, look for an exact match or
  // a timeUnit-prefixed match (e.g. "yearmonthdate_timestamp" for field "timestamp")
  if (temporalField) {
    for (const key in value) {
      if (key === temporalField || key.endsWith(`_${temporalField}`)) {
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
): { start: Date; end: Date } | undefined {
  const checkAndCreateTimeRange = (
    arr: unknown,
  ): { start: Date; end: Date } | undefined => {
    if (Array.isArray(arr) && arr.length === 2) {
      const [start, end] = arr;
      return { start: new Date(start), end: new Date(end) };
    }
    return undefined;
  };

  // Handle raw [date1, date2] array emitted directly by the brush_ts signal.
  // In Vega-Lite 6, brush_end/brush_clear reference brush_ts which yields the
  // interval as a bare array rather than a keyed object.
  if (Array.isArray(value)) {
    return checkAndCreateTimeRange(value);
  }

  /**
   * Time range fields can be either 'ts' or end with '_ts'
   * We check for both cases and return a TimeRange if a valid array of two timestamps is found.
   */
  if (typeof value === "object" && value !== null) {
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

    // Fallback: check any key with a 2-element array (handles arbitrary field names
    // from Vega-Lite brush selections where the key is the actual field name)
    for (const key in value) {
      const timeRange = checkAndCreateTimeRange(value[key]);
      if (timeRange) return timeRange;
    }
  }
  return undefined;
}
