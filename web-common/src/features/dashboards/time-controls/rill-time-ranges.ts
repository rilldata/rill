import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils";
import { get } from "svelte/store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  getQueryServiceMetricsViewTimeRangesQueryKey,
  queryServiceMetricsViewTimeRanges,
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export async function fetchTimeRanges(
  exploreSpec: V1ExploreSpec,
  defaultPreset: V1ExplorePreset,
) {
  const rillTimes = dedupe(
    [
      ...(defaultPreset.timeRange ? [defaultPreset.timeRange] : []),
      ...(exploreSpec.timeRanges?.length
        ? exploreSpec.timeRanges.map((t) => t.range!)
        : []),
    ].map((tr) => {
      try {
        const rillTime = parseRillTime(tr);
        if (defaultPreset.timezone) {
          rillTime.addTimezone(defaultPreset.timezone);
        }
        return rillTime.toString();
      } catch {
        return tr;
      }
    }),
  );
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
