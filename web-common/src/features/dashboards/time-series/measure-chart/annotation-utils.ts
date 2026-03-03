import type { ChartScales, ChartConfig, TimeSeriesPoint } from "./types";
import type {
  V1MetricsViewAnnotationsResponseAnnotation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
import { dateToIndex } from "./utils";
import type { DateTime } from "luxon";

export type Annotation = V1MetricsViewAnnotationsResponseAnnotation & {
  startTime: DateTime;
  endTime?: DateTime;
  formattedTimeOrRange: string;
};

export type AnnotationGroup = {
  items: Annotation[];
  key: string;
  index: number;
  left: number;
  right: number;
  top: number;
  bottom: number;
  hasRange: boolean;
};

export const AnnotationWidth = 10;
export const AnnotationHeight = 10;

/**
 * Group annotations by time grain bucket, then compute pixel positions.
 * All annotations whose startTime truncates to the same grain boundary
 * (in the given timezone) are grouped together.
 */
export function groupAnnotations(
  annotations: Annotation[],
  scales: ChartScales,
  data: TimeSeriesPoint[],
  config: ChartConfig,
  timeGrain: V1TimeGrain | undefined,
): AnnotationGroup[] {
  if (annotations.length === 0 || data.length === 0) return [];

  const unit = timeGrain ? V1TimeGrainToDateTimeUnit[timeGrain] : "day";
  const diamondY =
    config.plotBounds.top + config.plotBounds.height - AnnotationHeight;

  // Bucket annotations by their grain-truncated start time.
  // Store the truncated millis so the visibility check and positioning
  // use the bucket boundary (not the raw annotation time).
  const buckets = new Map<
    string,
    { annotations: Annotation[]; hasRange: boolean; bucketMs: number }
  >();

  for (const a of annotations) {
    const bucketStart = a.startTime.startOf(unit);
    const key =
      bucketStart.toISO() ??
      a.startTime.toISO() ??
      String(a.startTime.toMillis());

    let bucket = buckets.get(key);
    if (!bucket) {
      bucket = {
        annotations: [],
        hasRange: false,
        bucketMs: bucketStart.toMillis(),
      };
      buckets.set(key, bucket);
    }
    bucket.annotations.push(a);
    if (a.endTime) bucket.hasRange = true;
  }

  // Convert buckets to groups with pixel positions
  const groups: AnnotationGroup[] = [];

  // Data time range — annotations outside this are not visible
  const dataStartMs = data[0].ts.toMillis();
  const dataEndMs = data[data.length - 1].ts.toMillis();

  for (const [bucketKey, bucket] of buckets) {
    // Use the grain-truncated bucket time for visibility and positioning.
    // This ensures e.g. an hour annotation at 06:00 snaps to its day bucket
    // at 00:00, which aligns with the day-grain data point grid.
    if (bucket.bucketMs < dataStartMs || bucket.bucketMs > dataEndMs) continue;

    const startIdx = dateToIndex(data, bucket.bucketMs);
    if (startIdx === null) continue;

    const left = scales.x(startIdx);

    // Compute right edge from the widest range annotation in the bucket
    let right = left + AnnotationWidth;
    for (const a of bucket.annotations) {
      if (a.endTime) {
        const endIdx = dateToIndex(data, a.endTime.toMillis());
        if (endIdx !== null) {
          right = Math.max(right, scales.x(endIdx));
        }
      }
    }

    // Filter out-of-bounds groups
    if (
      left < config.plotBounds.left ||
      left > config.plotBounds.left + config.plotBounds.width
    ) {
      continue;
    }

    groups.push({
      items: bucket.annotations,
      key: bucketKey,
      index: startIdx,
      left,
      right: Math.min(right, config.plotBounds.left + config.plotBounds.width),
      top: diamondY,
      bottom: diamondY + AnnotationHeight,
      hasRange: bucket.hasRange,
    });
  }

  // Sort by x position
  groups.sort((a, b) => a.left - b.left);

  return groups;
}

export function findHoveredGroup(
  groups: AnnotationGroup[],
  mouseX: number,
  mouseY: number,
): AnnotationGroup | null {
  for (const group of groups) {
    if (
      mouseY >= group.top - 2 &&
      mouseY <= group.bottom + 2 &&
      mouseX >= group.left - AnnotationWidth / 2 &&
      mouseX <= group.left + AnnotationWidth / 2 + 2
    ) {
      return group;
    }
  }
  return null;
}
