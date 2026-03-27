import type { V1MetricsViewTimeSeriesResponse } from "@rilldata/web-common/runtime-client";
import type { TimeSeriesPoint } from "./types";
import { DateTime } from "luxon";

/**
 * Transform raw API time series data to typed TimeSeriesPoint[].
 * Minimal processing: just extract ts, value, and comparison fields.
 * No intermediate position computation — rendering uses indices directly.
 */
export function transformTimeSeriesData(
  primary: V1MetricsViewTimeSeriesResponse["data"],
  comparison: V1MetricsViewTimeSeriesResponse["data"] | undefined,
  measureName: string,
  timezone: string,
): TimeSeriesPoint[] {
  if (!primary) return [];

  return primary.map((originalPt, i) => {
    const comparisonPt = comparison?.[i];

    if (!originalPt?.ts) {
      return { ts: DateTime.invalid("Invalid timestamp"), value: null };
    }

    const ts = DateTime.fromISO(originalPt.ts, { zone: timezone });

    if (!ts.isValid) {
      return { ts: DateTime.invalid("Invalid timestamp"), value: null };
    }

    const rawValue = originalPt.records?.[measureName];
    const value =
      rawValue !== null && rawValue !== undefined ? Number(rawValue) : null;

    let comparisonValue: number | null | undefined = undefined;
    let comparisonTs: DateTime | undefined = undefined;

    if (comparisonPt?.ts) {
      const rawCompValue = comparisonPt.records?.[measureName];
      comparisonValue =
        rawCompValue !== null && rawCompValue !== undefined
          ? Number(rawCompValue)
          : null;
      comparisonTs = DateTime.fromISO(comparisonPt.ts, { zone: timezone });
    }

    return { ts, value, comparisonValue, comparisonTs };
  });
}
