import type { V1MetricsViewTimeRangeResponse } from "@rilldata/web-common/runtime-client";
import { createQueryServiceMetricsViewTimeRange } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";

export function createTimeRangeSummary(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewTimeRangeResponse, Error> {
  return derived(
    [ctx.metricsViewName, ctx.validSpecStore],
    ([metricsViewName, validSpec], set) =>
      createQueryServiceMetricsViewTimeRange(
        ctx.runtimeClient,
        {
          metricsViewName,
        },
        {
          query: {
            enabled:
              !validSpec.error && !!validSpec.data?.metricsView?.timeDimension,
          },
        },
        ctx.queryClient,
      ).subscribe(set),
  );
}
