import { bisector } from "d3-array";
import type { DateTimeUnit } from "luxon";
import type {
  TimeSeriesPoint,
  DimensionSeriesData,
  BisectedPoint,
} from "./types";
import {
  roundDownToTimeUnit,
  roundToNearestTimeUnit,
} from "../round-to-nearest-time-unit";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";

/**
 * D3 bisector for TimeSeriesPoint arrays, using the ts field.
 */
export const timeSeriesBisector = bisector<TimeSeriesPoint, Date>((d) => d.ts);

/**
 * Get the time grain label from V1TimeGrain.
 */
export function getTimeGrainLabel(timeGrain: V1TimeGrain): DateTimeUnit {
  return TIME_GRAIN[timeGrain]?.label as DateTimeUnit;
}

/**
 * Find the nearest point in the data array to the given date.
 * Uses optional time grain rounding for stable tooltip behavior.
 */
export function bisectTimeSeriesPoint(
  data: TimeSeriesPoint[],
  date: Date,
  timeGrain: V1TimeGrain,
  roundStrategy: "nearest" | "down" = "down",
): BisectedPoint {
  if (!data.length || !date) {
    return { point: null, index: -1, roundedTs: null };
  }

  const grainLabel = getTimeGrainLabel(timeGrain);
  if (!grainLabel) {
    return { point: null, index: -1, roundedTs: null };
  }

  // Round to time grain for stable tooltip
  const roundedTs =
    roundStrategy === "down"
      ? roundDownToTimeUnit(date, grainLabel)
      : roundToNearestTimeUnit(date, grainLabel);

  // Find nearest point using bisector center
  const index = timeSeriesBisector.center(data, roundedTs);

  // Clamp index to valid range
  const clampedIndex = Math.max(0, Math.min(data.length - 1, index));
  const point = data[clampedIndex] ?? null;

  return { point, index: clampedIndex, roundedTs };
}

/**
 * Find values for all dimension series at a given time.
 * Returns a map from dimension value to the data point.
 */
export function bisectDimensionData(
  dimensionData: DimensionSeriesData[],
  date: Date,
  timeGrain: V1TimeGrain,
): Map<
  string | null,
  { value: number | null; color: string; point: TimeSeriesPoint | null }
> {
  const results = new Map<
    string | null,
    { value: number | null; color: string; point: TimeSeriesPoint | null }
  >();

  for (const dim of dimensionData) {
    const { point } = bisectTimeSeriesPoint(dim.data, date, timeGrain);
    results.set(dim.dimensionValue, {
      value: point?.value ?? null,
      color: dim.color,
      point,
    });
  }

  return results;
}

/**
 * Check if a date is within the data's time range.
 */
export function isDateInDataRange(
  date: Date,
  data: TimeSeriesPoint[],
): boolean {
  if (!data.length) return false;

  const firstTs = data[0].ts;
  const lastTs = data[data.length - 1].ts;

  return date >= firstTs && date <= lastTs;
}

/**
 * Find the closest data point to the given screen X coordinate.
 * Useful for mouse interactions.
 */
export function bisectByScreenX(
  screenX: number,
  data: TimeSeriesPoint[],
  xScale: (date: Date) => number,
  timeGrain: V1TimeGrain,
): BisectedPoint {
  if (!data.length) {
    return { point: null, index: -1, roundedTs: null };
  }

  // Get the time range
  const firstTs = data[0].ts;
  const lastTs = data[data.length - 1].ts;

  // Convert screen X to date via inverse of scale
  // We need to approximate this since we don't have the inverse scale here
  const firstX = xScale(firstTs);
  const lastX = xScale(lastTs);

  // Linear interpolation to get approximate date
  const ratio = (screenX - firstX) / (lastX - firstX);
  const timeRange = lastTs.getTime() - firstTs.getTime();
  const approximateTime = new Date(firstTs.getTime() + ratio * timeRange);

  return bisectTimeSeriesPoint(data, approximateTime, timeGrain);
}

/**
 * Compute line segments from data, handling gaps (null values).
 * Returns an array of segments, where each segment is a contiguous
 * run of non-null data points.
 */
export function computeLineSegments(
  data: TimeSeriesPoint[],
): TimeSeriesPoint[][] {
  const segments: TimeSeriesPoint[][] = [];
  let currentSegment: TimeSeriesPoint[] = [];

  for (const point of data) {
    if (point.value !== null) {
      currentSegment.push(point);
    } else if (currentSegment.length > 0) {
      segments.push(currentSegment);
      currentSegment = [];
    }
  }

  // Don't forget the last segment
  if (currentSegment.length > 0) {
    segments.push(currentSegment);
  }

  return segments;
}

/**
 * Find singleton points (segments with only one point).
 * These need to be rendered as circles instead of lines.
 */
export function findSingletonPoints(
  data: TimeSeriesPoint[],
): TimeSeriesPoint[] {
  const segments = computeLineSegments(data);
  return segments
    .filter((segment) => segment.length === 1)
    .map((segment) => segment[0]);
}
