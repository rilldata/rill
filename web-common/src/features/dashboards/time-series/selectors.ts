import type { Annotation } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import {
  getQueryServiceMetricsViewAnnotationsQueryOptions,
  type V1MetricsViewAnnotationsResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import { createQueries } from "@tanstack/svelte-query";
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
        .map((r) => r.data?.data?.map(convertDateItemToAnnotation) ?? [])
        .flat();
      annotations.sort((a, b) => a.time.getTime() - b.time.getTime());
      return annotations;
    },
  });
}

function convertDateItemToAnnotation(
  item: V1MetricsViewAnnotationsResponseDataItem,
) {
  return <Annotation>{
    ...item,
    time: new Date(item.time as string),
    time_end: item.time_end ? new Date(item.time_end as string) : undefined,
  };
}
