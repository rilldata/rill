import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import {
  getOffset,
  getStartOfPeriod,
} from "@rilldata/web-common/lib/time/transforms";
import {
  type DashboardTimeControls,
  Period,
  TimeOffsetType,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  getQueryServiceMetricsViewAnnotationsQueryOptions,
  type V1MetricsViewAnnotationsResponseAnnotation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { createQuery } from "@tanstack/svelte-query";
import { DateTime, Interval } from "luxon";
import { derived, type Readable } from "svelte/store";

export function getAnnotationsForMeasure({
  instanceId,
  exploreName,
  measureName,
  selectedTimeRange,
}: {
  instanceId: string;
  exploreName: string;
  measureName: string;
  selectedTimeRange: DashboardTimeControls | undefined;
}): Readable<Annotation[]> {
  const exploreValidSpec = useExploreValidSpec(instanceId, exploreName);
  const selectedPeriod = TIME_GRAIN[selectedTimeRange?.interval ?? ""]
    ?.duration as Period | undefined;

  const annotationsQueryOptions = derived(
    exploreValidSpec,
    (exploreValidSpec) => {
      const metricsViewSpec = exploreValidSpec.data?.metricsView;
      const exploreSpec = exploreValidSpec.data?.explore;
      const metricsViewName = exploreSpec?.metricsView ?? "";

      return getQueryServiceMetricsViewAnnotationsQueryOptions(
        instanceId,
        metricsViewName,
        {
          timeRange: {
            start: selectedTimeRange?.start.toISOString(),
            end: selectedTimeRange?.end.toISOString(),
          },
          timeGrain: selectedTimeRange?.interval,
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
          getLocalIANA(), // Use system timezone for annotations similar to chart labels
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
  timezone: string,
) {
  let startTime = new Date(annotation.time as string);
  let endTime = annotation.timeEnd ? new Date(annotation.timeEnd) : undefined;

  // Only truncate start and ceil end when there is a grain column in the annotation.
  if (period && annotation.duration) {
    startTime = getStartOfPeriod(startTime, period, timezone);
    if (endTime) {
      endTime = getOffset(endTime, period, TimeOffsetType.ADD, timezone);
      endTime = getStartOfPeriod(endTime, period, timezone);
    }
  }

  const formattedTimeOrRange = prettyFormatTimeRange(
    Interval.fromDateTimes(
      DateTime.fromJSDate(startTime).setZone(timezone),
      DateTime.fromJSDate(endTime ?? startTime).setZone(timezone),
    ),
    selectedTimeGrain,
  );

  return <Annotation>{
    ...annotation,
    startTime,
    endTime,
    formattedTimeOrRange,
  };
}
