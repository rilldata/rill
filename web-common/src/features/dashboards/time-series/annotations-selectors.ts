import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { prettyFormatTimeRangeV2 } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
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
          selectedTimezone,
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
  selectedTimezone: string,
) {
  let startTime = new Date(annotation.time as string);
  let endTime = annotation.timeEnd ? new Date(annotation.timeEnd) : undefined;

  // Only truncate start and ceil end when there is a grain column in the annotation.
  if (period && annotation.duration) {
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

  const formattedTimeOrRange = prettyFormatTimeRangeV2(
    startTime,
    endTime ?? startTime,
    selectedTimeGrain,
    selectedTimezone,
  );

  return <Annotation>{
    ...annotation,
    startTime,
    endTime,
    formattedTimeOrRange,
  };
}
