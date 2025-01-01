import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils";
import {
  createQueryServiceMetricsViewResolveTimeRanges,
  type V1MetricsViewResolveTimeRangesResponse,
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

      createQueryServiceMetricsViewResolveTimeRanges(
        get(runtime).instanceId,
        explore.metricsView!,
        {
          rillTimes,
        },
      ).subscribe(set);
    },
  ) as CreateQueryResult<V1MetricsViewResolveTimeRangesResponse>;
}
