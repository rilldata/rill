import { page } from "$app/stores";
import {
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  getAdminServiceListBookmarksQueryOptions,
  type V1Bookmark,
  type V1ListBookmarksResponse,
} from "@rilldata/web-admin/client";
import {
  categorizeBookmarks,
  parseBookmarks,
} from "@rilldata/web-admin/features/bookmarks/utils.ts";
import {
  getProjectIdQueryOptions,
  type OrgAndProjectNameStore,
} from "@rilldata/web-admin/features/projects/selectors.ts";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState.ts";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getBookmarksQueryOptions(
  orgAndProjectNameStore: OrgAndProjectNameStore,
  resourceKind: ResourceKind, // This doesnt change
  resourceNameStore: Readable<string>,
) {
  const projectIdQuery = createQuery(
    getProjectIdQueryOptions(orgAndProjectNameStore),
  );

  return derived(
    [createAdminServiceGetCurrentUser(), projectIdQuery, resourceNameStore],
    ([userResp, projectIdQueryResp, resourceName]) => {
      const hasUser = userResp.data?.user;
      const projectId = projectIdQueryResp.data;

      return getAdminServiceListBookmarksQueryOptions(
        {
          projectId,
          resourceKind,
          resourceName,
        },
        {
          query: {
            enabled: hasUser && !!projectId,
          },
        },
      );
    },
  );
}

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
  // TODO: refactor to use query options and a stable query.
  return getCompoundQuery(
    [
      getBookmarks(projectId, ResourceKind.Explore, exploreName),
      useExploreValidSpec(instanceId, exploreName),
      useMetricsViewTimeRange(instanceId, metricsViewName),
    ],
    ([bookmarksResp, exploreSpecResp, timeRangeResp]) => {
      const homeBookmark = bookmarksResp?.bookmarks?.find(isHomeBookmark);
      if (!homeBookmark) return null;

      const metricsViewSpec = exploreSpecResp?.metricsView ?? {};
      const exploreSpec = exploreSpecResp?.explore ?? {};

      if (homeBookmark.data) {
        // Legacy bookmark data stored in proto format.
        const exploreStateFromLegacyProto = getDashboardStateFromUrl(
          homeBookmark.data,
          metricsViewSpec,
          exploreSpec,
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

export function getCanvasCategorisedBookmarks(
  orgAndProjectNameStore: OrgAndProjectNameStore,
  canvasNameStore: Readable<string>,
) {
  const bookmarksQuery = createQuery(
    getBookmarksQueryOptions(
      orgAndProjectNameStore,
      ResourceKind.Canvas,
      canvasNameStore,
    ),
  );

  return derived([bookmarksQuery, page], ([bookmarksResp, pageState]) => {
    const bookmarks = bookmarksResp.data?.bookmarks ?? [];
    const parsedBookmarks = parseBookmarks(
      bookmarks,
      pageState.url.searchParams,
    );
    const categorizedBookmarks = categorizeBookmarks(parsedBookmarks);
    return {
      data: {
        bookmarks,
        categorizedBookmarks,
      },
      isPending: bookmarksResp.isPending,
      error: bookmarksResp.error,
    };
  });
}
