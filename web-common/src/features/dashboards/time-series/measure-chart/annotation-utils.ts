import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations";
import type { ChartScales, ChartConfig, TimeSeriesPoint } from "./types";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
import { DateTime } from "luxon";
import { dateToIndex } from "./utils";

export type AnnotationGroup = {
  items: Annotation[];
  /** Data index this group maps to */
  index: number;
  /** Pixel x of the group diamond */
  left: number;
  /** Pixel x of rightmost annotation end (for range annotations) */
  right: number;
  /** Pixel y top of diamond area */
  top: number;
  /** Pixel y bottom of diamond area */
  bottom: number;
  /** Whether any item in the group has a range (endTime) */
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
  timeZone: string,
): AnnotationGroup[] {
  if (annotations.length === 0 || data.length === 0) return [];

  const unit = timeGrain ? V1TimeGrainToDateTimeUnit[timeGrain] : "day";
  const diamondY =
    config.plotBounds.top + config.plotBounds.height - AnnotationHeight;

  // Bucket annotations by their truncated grain key
  const buckets = new Map<
    string,
    { annotations: Annotation[]; hasRange: boolean }
  >();

  for (const a of annotations) {
    const dt = DateTime.fromJSDate(a.startTime, { zone: timeZone });
    const key =
      dt.startOf(unit).toISO() ?? dt.toISO() ?? String(a.startTime.getTime());

    let bucket = buckets.get(key);
    if (!bucket) {
      bucket = { annotations: [], hasRange: false };
      buckets.set(key, bucket);
    }
    bucket.annotations.push(a);
    if (a.endTime) bucket.hasRange = true;
  }

  // Convert buckets to groups with pixel positions
  const groups: AnnotationGroup[] = [];

  for (const [, bucket] of buckets) {
    // Use the first annotation's startTime for positioning
    const startIdx = dateToIndex(
      data,
      bucket.annotations[0].startTime.getTime(),
    );
    if (startIdx === null) continue;

    const left = scales.x(startIdx);

    // Compute right edge from the widest range annotation in the bucket
    let right = left + AnnotationWidth;
    for (const a of bucket.annotations) {
      if (a.endTime) {
        const endIdx = dateToIndex(data, a.endTime.getTime());
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
