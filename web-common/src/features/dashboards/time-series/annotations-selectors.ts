import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import {
  getEndOfPeriod,
  getStartOfPeriod,
} from "@rilldata/web-common/lib/time/transforms";
import {
  type DashboardTimeControls,
  Period,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  getQueryServiceMetricsViewAnnotationsQueryOptions,
  type V1MetricsViewAnnotationsResponseDataItem,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { createQueries } from "@tanstack/svelte-query";
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
              convertDateItemToAnnotation(
                a,
                selectedTimeRange?.interval ??
                  V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
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

function convertDateItemToAnnotation(
  item: V1MetricsViewAnnotationsResponseDataItem,
  timeGrain: V1TimeGrain,
  selectedTimezone: string,
) {
  const startTime = new Date(item.time as string);
  let truncatedStartTime = startTime;
  const endTime = item.time_end ? new Date(item.time_end as string) : undefined;
  let truncatedEndTime = endTime;

  const period: Period | undefined = TIME_GRAIN[timeGrain]?.duration;
  if (period) {
    truncatedStartTime = getStartOfPeriod(startTime, period, selectedTimezone);
    if (endTime) {
      truncatedEndTime = getEndOfPeriod(endTime, period, selectedTimezone);
    }
  }

  return <Annotation>{
    ...item,
    startTime,
    truncatedStartTime,
    endTime,
    truncatedEndTime,
  };
}
