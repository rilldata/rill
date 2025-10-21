import {
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  type V1Bookmark,
  type V1ListBookmarksResponse,
} from "@rilldata/web-admin/client";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState.ts";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewSchema,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1StructType,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getBookmarks(
  projectId: string,
  resourceKind: ResourceKind,
  resourceName: string,
) {
  return derived(createAdminServiceGetCurrentUser(), (userResp, set) =>
    createAdminServiceListBookmarks(
      {
        projectId,
        resourceKind,
        resourceName,
      },
      {
        query: {
          enabled: !!userResp.data?.user && !!projectId,
        },
      },
      queryClient,
    ).subscribe(set),
  ) as CreateQueryResult<V1ListBookmarksResponse, HTTPError>;
}

export function isHomeBookmark(bookmark: V1Bookmark) {
  return Boolean(bookmark.default);
}

export function getHomeBookmarkExploreState(
  projectId: string,
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
): CompoundQueryResult<Partial<ExploreState> | null> {
  return getCompoundQuery(
    [
      getBookmarks(projectId, ResourceKind.Explore, exploreName),
      useExploreValidSpec(instanceId, exploreName),
      createQueryServiceMetricsViewSchema(instanceId, metricsViewName),
      useMetricsViewTimeRange(instanceId, metricsViewName),
    ],
    ([bookmarksResp, exploreSpecResp, schemaResp, timeRangeResp]) => {
      const homeBookmark = bookmarksResp?.bookmarks?.find(isHomeBookmark);
      if (!homeBookmark) return null;

      const metricsViewSpec = exploreSpecResp?.metricsView ?? {};
      const exploreSpec = exploreSpecResp?.explore ?? {};

      if (homeBookmark.data) {
        const exploreStateFromLegacyProto = getDashboardStateFromUrl(
          homeBookmark.data,
          metricsViewSpec,
          exploreSpec,
          schemaResp?.schema ?? {},
        );
        return exploreStateFromLegacyProto;
      }

      const explorePreset = getDefaultExplorePreset(
        exploreSpec,
        metricsViewSpec,
        timeRangeResp?.timeRangeSummary,
      );

      const { partialExploreState: exploreStateFromHomeBookmark } =
        convertURLSearchParamsToExploreState(
          new URLSearchParams(homeBookmark.urlSearch ?? ""),
          metricsViewSpec,
          exploreSpec,
          explorePreset,
        );
      return exploreStateFromHomeBookmark;
    },
  );
}

export function exploreBookmarkDataTransformer({
  data,
  metricsViewSpec,
  exploreSpec,
  schema,
  timeRangeSummary,
}: {
  data: string;
  metricsViewSpec: V1MetricsViewSpec;
  exploreSpec: V1ExploreSpec;
  schema: V1StructType;
  timeRangeSummary: V1TimeRangeSummary | undefined;
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
