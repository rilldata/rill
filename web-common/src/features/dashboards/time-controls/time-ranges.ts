import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewTimeRanges,
  getQueryServiceMetricsViewTimeRangesQueryKey,
  queryServiceMetricsViewTimeRanges,
  type V1ExploreSpec,
  type V1MetricsViewTimeRangesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export function getTimeRanges(exploreName: string) {
  return derived(
    [useExploreValidSpec(get(runtime).instanceId, exploreName)],
    ([validSpecResp], set) => {
      if (!validSpecResp.data?.explore) {
        return;
      }

      const explore = validSpecResp.data.explore;
      const defaultPreset = explore.defaultPreset ?? {};
      const rillTimes = dedupe([
        ...(defaultPreset.timeRange ? [defaultPreset.timeRange] : []),
        ...(explore.timeRanges?.length
          ? explore.timeRanges.map((t) => t.range!)
          : []),
      ]);

      createQueryServiceMetricsViewTimeRanges(
        get(runtime).instanceId,
        explore.metricsView!,
        {
          expressions: rillTimes,
        },
      ).subscribe(set);
    },
  ) as CreateQueryResult<V1MetricsViewTimeRangesResponse>;
}

export async function fetchTimeRanges(exploreSpec: V1ExploreSpec) {
  const defaultPreset = exploreSpec.defaultPreset ?? {};
  const rillTimes = dedupe([
    ...(defaultPreset.timeRange ? [defaultPreset.timeRange] : []),
    ...(exploreSpec.timeRanges?.length
      ? exploreSpec.timeRanges.map((t) => t.range!)
      : []),
  ]);
  const instanceId = get(runtime).instanceId;
  const metricsViewName = exploreSpec.metricsView!;

  const timeRangesResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewTimeRangesQueryKey(
      instanceId,
      metricsViewName,
      { expressions: rillTimes },
    ),
    queryFn: () =>
      queryServiceMetricsViewTimeRanges(instanceId, metricsViewName, {
        expressions: rillTimes,
      }),
  });
  return timeRangesResp.timeRanges ?? [];
}
