import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import {
  type DashboardTimeControls,
  Period,
  TimeUnit,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  getQueryServiceMetricsViewAnnotationsQueryOptions,
  type V1MetricsViewAnnotationsResponseAnnotation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQuery } from "@tanstack/svelte-query";
import { DateTime, Interval } from "luxon";
import { derived, type Readable } from "svelte/store";

export function getAnnotationsForMeasure({
  client,
  exploreName,
  measureName,
  selectedTimeRange,
  dashboardTimezone,
}: {
  client: RuntimeClient;
  exploreName: string;
  measureName: string;
  selectedTimeRange: DashboardTimeControls | undefined;
  dashboardTimezone: string;
}): Readable<Annotation[]> {
  const exploreValidSpec = useExploreValidSpec(client, exploreName);
  const selectedPeriod = TIME_GRAIN[selectedTimeRange?.interval ?? ""]
    ?.duration as Period | undefined;

  const annotationsQueryOptions = derived(
    exploreValidSpec,
    (exploreValidSpec) => {
      const metricsViewSpec = exploreValidSpec.data?.metricsView;
      const exploreSpec = exploreValidSpec.data?.explore;
      const metricsViewName = exploreSpec?.metricsView ?? "";

      return getQueryServiceMetricsViewAnnotationsQueryOptions(
        client,
        {
          metricsViewName,
          timeRange: {
            start: selectedTimeRange?.start.toISOString() as any,
            end: selectedTimeRange?.end.toISOString() as any,
          },
          timeGrain: selectedTimeRange?.interval as any,
          measures: [measureName],
        },
        {
          query: {
            enabled:
              !!metricsViewSpec?.annotations?.length &&
              !!metricsViewName &&
              !!selectedTimeRange,
          },
        },
      );
    },
  );

  const annotationsQuery = createQuery(annotationsQueryOptions);

  return derived(annotationsQuery, (annotationsQuery) => {
    const annotations =
      annotationsQuery.data?.rows?.map((a) =>
        convertV1AnnotationsResponseItemToAnnotation(
          a,
          selectedPeriod,
          selectedTimeRange?.interval ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,

          dashboardTimezone,
        ),
      ) ?? [];
    annotations.sort((a, b) => a.startTime.getTime() - b.startTime.getTime());
    return annotations;
  });
}

function convertV1AnnotationsResponseItemToAnnotation(
  annotation: V1MetricsViewAnnotationsResponseAnnotation,
  period: Period | undefined,
  selectedTimeGrain: V1TimeGrain,
  dashboardTimezone: string,
) {
  const localTimezone = getLocalIANA();

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
    startTime: startTime
      .setZone(localTimezone, { keepLocalTime: true })
      .toJSDate(),
    endTime: endTime
      ?.setZone(localTimezone, { keepLocalTime: true })
      .toJSDate(),
    formattedTimeOrRange,
  };
}
