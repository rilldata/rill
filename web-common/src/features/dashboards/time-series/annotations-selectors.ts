import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { Period, TimeUnit } from "@rilldata/web-common/lib/time/types.ts";
import {
  createQueryServiceMetricsViewAnnotations,
  type V1MetricsViewAnnotationsResponseAnnotation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { keepPreviousData } from "@tanstack/svelte-query";

/**
 * Creates a query that fetches annotations for a measure and transforms
 * the raw response rows into sorted Annotation objects via `select`.
 */
export function createAnnotationsQuery(
  instanceId: string,
  metricsViewName: string,
  measureName: string,
  timeDimension: string | undefined,
  timeStart: string | undefined,
  timeEnd: string | undefined,
  timeGranularity: V1TimeGrain | undefined,
  timeZone: string,
  enabled: boolean,
) {
  const period = getPeriodFromTimeGrain(timeGranularity);
  const grain = timeGranularity ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

  return createQueryServiceMetricsViewAnnotations(
    instanceId,
    metricsViewName,
    {
      timeRange: { start: timeStart, end: timeEnd, timeDimension },
      timeGrain: timeGranularity,
      measures: [measureName],
    },
    {
      query: {
        select: (data) => {
          const rows = data.rows;
          if (!rows?.length) return [] as Annotation[];
          const list = rows.map((a) =>
            convertV1AnnotationsResponseItemToAnnotation(
              a,
              period,
              grain,
              timeZone,
            ),
          );
          list.sort((a, b) => a.startTime.toMillis() - b.startTime.toMillis());
          return list;
        },
        enabled,
        placeholderData: keepPreviousData,
        refetchOnMount: false,
      },
    },
  );
}

function convertV1AnnotationsResponseItemToAnnotation(
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

function getPeriodFromTimeGrain(
  timeGrain: V1TimeGrain | string | undefined,
): Period | undefined {
  return TIME_GRAIN[timeGrain ?? ""]?.duration as Period | undefined;
}
