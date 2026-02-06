import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  Period,
  TimeUnit,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  type V1MetricsViewAnnotationsResponseAnnotation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";

export function convertV1AnnotationsResponseItemToAnnotation(
  annotation: V1MetricsViewAnnotationsResponseAnnotation,
  period: Period | undefined,
  selectedTimeGrain: V1TimeGrain,
  dashboardTimezone: string,
): Annotation {
  let startTime = DateTime.fromISO(annotation.time as string, {
    zone: dashboardTimezone,
  });
  let endTime = annotation.timeEnd
    ? DateTime.fromISO(annotation.timeEnd, {
        zone: dashboardTimezone,
      })
    : undefined;

  // Only truncate start and ceil end when there is a grain column in the annotation.
  if (period && annotation.duration) {
    startTime = startTime.startOf(TimeUnit[period]);
    if (endTime) {
      endTime = startTime
        .plus({ [TimeUnit[period]]: 1 })
        .startOf(TimeUnit[period]);
    }
  }

  const formattedTimeOrRange = prettyFormatTimeRange(
    Interval.fromDateTimes(startTime, endTime ?? startTime),
    selectedTimeGrain,
  );

  return <Annotation>{
    ...annotation,
    startTime,
    endTime,
    formattedTimeOrRange,
  };
}

/**
 * Derive the Period from a time grain string, used for annotation truncation.
 */
export function getPeriodFromTimeGrain(
  timeGrain: V1TimeGrain | string | undefined,
): Period | undefined {
  return TIME_GRAIN[timeGrain ?? ""]?.duration as Period | undefined;
}
