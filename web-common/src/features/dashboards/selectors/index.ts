import {
  type V1MetricsViewTimeRangeResponse,
  createQueryServiceMetricsViewSchema,
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewSchemaResponse,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

export function createTimeRangeSummary(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewTimeRangeResponse, HTTPError> {
  return derived(
    [ctx.runtime, ctx.metricsViewName, ctx.validSpecStore],
    ([runtime, metricsViewName, validSpec], set) =>
      createQueryServiceMetricsViewTimeRange(
        runtime.instanceId,
        metricsViewName,
        {},
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

export function createMetricsViewSchema(
  ctx: StateManagers,
): CreateQueryResult<V1MetricsViewSchemaResponse, HTTPError> {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricsViewName], set) =>
      createQueryServiceMetricsViewSchema(
        runtime.instanceId,
        metricsViewName,
        {},
        {},
        ctx.queryClient,
      ).subscribe(set),
  );
}
