import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getRillDefaultExploreUrlParams(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const rillDefaultExploreState = getRillDefaultExploreState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  const timeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    rillDefaultExploreState,
  );
  return convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    rillDefaultExploreState,
    timeControlState,
  );
}

/**
 * Rill default explore url params with a stable non-reactive query object.
 * All reactivity will instead be in the query options.
 *
 * Uses {@link getExploreValidSpecQueryOptions} and {@link getMetricsViewTimeRangeFromExploreQueryOptions} for reactive query options.
 */
export function createRillDefaultExploreUrlParamsV2(
  client: RuntimeClient,
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(client, exploreNameStore),
  );
  const timeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(client, exploreNameStore),
  );

  return derived(
    [validSpecQuery, timeRangeQuery],
    ([validSpecResp, timeRangeResp]) => {
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const timeRangeSummary = timeRangeResp.data?.timeRangeSummary;

      if (
        !metricsViewSpec ||
        !exploreSpec ||
        // safeguard to make sure time range summary is loaded for metrics view with time dimension
        (metricsViewSpec.timeDimension && !timeRangeSummary)
      ) {
        return undefined;
      }

      return getRillDefaultExploreUrlParams(
        metricsViewSpec,
        exploreSpec,
        timeRangeSummary,
      );
    },
  );
}
