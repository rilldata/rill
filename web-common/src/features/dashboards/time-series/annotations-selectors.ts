import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import {
  formatRange,
  formatTime,
} from "@rilldata/web-common/features/dashboards/time-controls/range-formatting.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
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
} from "@rilldata/web-common/runtime-client";
import { createQueries } from "@tanstack/svelte-query";
import { DateTime, Interval } from "luxon";
import { derived, type Readable } from "svelte/store";

export function getAnnotationsForMeasure({
  instanceId,
  exploreName,
  measureName,
  selectedTimeRange,
  selectedTimezone,
}: {
  instanceId: string;
  exploreName: string;
  measureName: string;
  selectedTimeRange: DashboardTimeControls | undefined;
  selectedTimezone: string;
}): Readable<Annotation[]> {
  const exploreValidSpec = useExploreValidSpec(instanceId, exploreName);
  const selectedPeriod = TIME_GRAIN[selectedTimeRange?.interval ?? ""]
    ?.duration as Period | undefined;

  const annotationsQueries = derived(exploreValidSpec, (exploreValidSpec) => {
    const metricsViewSpec = exploreValidSpec.data?.metricsView;
    const exploreSpec = exploreValidSpec.data?.explore;
    const metricsViewName = exploreSpec?.metricsView;
    const annotations = metricsViewSpec?.annotations;
    if (!metricsViewName || !annotations) return [];

    const annotationNames = annotations.filter((a) =>
      a.measures?.includes(measureName),
    );
    return annotationNames.map((annotation) =>
      getQueryServiceMetricsViewAnnotationsQueryOptions(
        instanceId,
        metricsViewName,
        annotation.name!,
        {
          timeRange: {
            start: selectedTimeRange?.start.toISOString(),
            end: selectedTimeRange?.end.toISOString(),
          },
          timeGrain: selectedTimeRange?.interval,
        },
        {
          query: {
            enabled: !!selectedTimeRange,
          },
        },
      ),
    );
  });

  return createQueries({
    queries: annotationsQueries,
    combine: (responses) => {
      const annotations = responses
        .map(
          (r) =>
            r.data?.data?.map((a) =>
              convertV1AnnotationsResponseItemToAnnotation(
                a,
                selectedPeriod,
                selectedTimezone,
              ),
            ) ?? [],
        )
        .flat();
      annotations.sort((a, b) => a.startTime.getTime() - b.startTime.getTime());
      return annotations;
    },
  });
}

function convertV1AnnotationsResponseItemToAnnotation(
  annotation: V1MetricsViewAnnotationsResponseAnnotation,
  period: Period | undefined,
  selectedTimezone: string,
) {
  let startTime = new Date(annotation.time as string);
  let endTime = annotation.timeEnd ? new Date(annotation.timeEnd) : undefined;

  // Only truncate start and ceil end when there is a grain column in the annotation.
  if (period && annotation.grain) {
    startTime = getStartOfPeriod(startTime, period, selectedTimezone);
    if (endTime) {
      endTime = getOffset(
        endTime,
        period,
        TimeOffsetType.ADD,
        selectedTimezone,
      );
      endTime = getStartOfPeriod(endTime, period, selectedTimezone);
    }
  }

  const formattedTimeOrRange = endTime
    ? formatRange(Interval.fromDateTimes(startTime, endTime))
    : formatTime(DateTime.fromJSDate(startTime));

  return <Annotation>{
    ...annotation,
    startTime,
    endTime,
    formattedTimeOrRange,
  };
}
