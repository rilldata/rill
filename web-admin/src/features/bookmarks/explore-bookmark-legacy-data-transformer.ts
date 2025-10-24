import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto.ts";
import {
  getMetricsViewSchemaOptions,
  getMetricsViewTimeRangeFromExploreQueryOptions,
} from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1StructType,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function createExploreBookmarkLegacyDataTransformer(
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );
  const timeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(exploreNameStore),
  );
  const schemaQuery = createQuery(
    getMetricsViewSchemaOptions(exploreNameStore),
  );

  return derived(
    [validSpecQuery, timeRangeQuery, schemaQuery],
    ([validSpecResp, timeRangeResp, schemaResp]) => {
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const timeRangeSummary = timeRangeResp.data?.timeRangeSummary;
      const schema = schemaResp.data?.schema ?? {};

      return (data: string) =>
        exploreBookmarkDataTransformer({
          data,
          metricsViewSpec,
          exploreSpec,
          timeRangeSummary,
          schema,
        });
    },
  );
}

export function exploreBookmarkDataTransformer({
  data,
  metricsViewSpec,
  exploreSpec,
  timeRangeSummary,
  schema,
}: {
  data: string;
  metricsViewSpec: V1MetricsViewSpec;
  exploreSpec: V1ExploreSpec;
  timeRangeSummary: V1TimeRangeSummary | undefined;
  schema: V1StructType;
}) {
  const exploreStateFromBookmark = getDashboardStateFromUrl(
    data,
    metricsViewSpec,
    exploreSpec,
    schema,
  );

  // We need to check if the bookmark's url is equal to current url or not to show an "active" state.
  // To avoid calculating it everytime we directly convert it to final url.
  const searchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    exploreStateFromBookmark,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
      exploreStateFromBookmark,
    ),
  );

  return "?" + searchParams.toString();
}
