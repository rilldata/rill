import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import {
  getExploreValidSpecQueryOptions,
  useExploreValidSpec,
} from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
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
    rillDefaultExploreState,
    timeControlState,
  );
}

export function createRillDefaultExploreUrlParams(
  validSpecQuery: ReturnType<typeof useExploreValidSpec>,
  fullTimeRangeQuery: CompoundQueryResult<V1MetricsViewTimeRangeResponse>,
) {
  return getCompoundQuery(
    [validSpecQuery, fullTimeRangeQuery],
    ([validSpecResp, metricsViewTimeRangeResp]) => {
      const metricsViewSpec = validSpecResp?.metricsView;
      const exploreSpec = validSpecResp?.explore;

      if (
        !metricsViewSpec ||
        !exploreSpec ||
        // safeguard to make sure time range summary is loaded for metrics view with time dimension
        (metricsViewSpec.timeDimension &&
          !metricsViewTimeRangeResp?.timeRangeSummary)
      ) {
        return undefined;
      }

      return getRillDefaultExploreUrlParams(
        metricsViewSpec,
        exploreSpec,
        metricsViewTimeRangeResp?.timeRangeSummary,
      );
    },
  );
}

/**
 * Version of createRillDefaultExploreUrlParams that is meant to have a stable non-reactive query object.
 * All reactivity will instead be in the query options.
 *
 * Uses {@link getExploreValidSpecQueryOptions} and {@link getMetricsViewTimeRangeFromExploreQueryOptions} for reactive query options.
 * TODO: replace {@link createRillDefaultExploreUrlParams} with this
 */
export function createRillDefaultExploreUrlParamsV2(
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );
  const timeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(exploreNameStore),
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
