import { page } from "$app/stores";
import {
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  getAdminServiceListBookmarksQueryKey,
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
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { HTTPError } from "@rilldata/web-common/lib/errors";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import { convertProtoToUrlParams } from "@rilldata/web-common/features/dashboards/proto-state/convert-proto-to-url-params.ts";
import { normalizeExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/merge-explore-url-params.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";

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
            enabled: Boolean(hasUser && !!projectId),
          },
        },
      );
    },
  );
}

/**
 * Query options for every bookmark in the project that is visible to the current user.
 * This powers the bookmark manager; the dashboard dropdowns use {@link getBookmarksQueryOptions} scoped to one resource.
 */
export function getProjectBookmarksQueryOptions(
  orgAndProjectNameStore: OrgAndProjectNameStore,
) {
  const projectIdQuery = createQuery(
    getProjectIdQueryOptions(orgAndProjectNameStore),
  );

  return derived(
    [createAdminServiceGetCurrentUser(), projectIdQuery],
    ([userResp, projectIdQueryResp]) => {
      const projectId = projectIdQueryResp.data;

      return getAdminServiceListBookmarksQueryOptions(
        { projectId },
        {
          query: {
            enabled: Boolean(userResp.data?.user && projectId),
          },
        },
      );
    },
  );
}

/**
 * Invalidates every bookmarks list query: the per-dashboard ones behind the bookmark dropdowns and the project-wide one behind the bookmark manager.
 * The generated query keys are `["/v1/users/bookmarks", params]`, so the param-less key prefix-matches all of them.
 * Inactive queries are refetched too: the global query client disables refetchOnMount,
 * so a list that is not on screen (e.g. the project home while on a dashboard) would otherwise show stale data when it remounts.
 */
export function invalidateBookmarkQueries() {
  return queryClient.invalidateQueries({
    queryKey: getAdminServiceListBookmarksQueryKey(),
    refetchType: "all",
  });
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

export function getHomeBookmarkUrlParams(
  projectId: string,
  exploreName: string,
): CompoundQueryResult<URLSearchParams | null> {
  // TODO: refactor to use query options and a stable query.
  return getCompoundQuery(
    [getBookmarks(projectId, ResourceKind.Explore, exploreName)],
    ([bookmarksResp]) => {
      const homeBookmark = bookmarksResp?.bookmarks?.find(isHomeBookmark);
      if (!homeBookmark) return null;
      return getUrlParamsFromHomeBookmark(homeBookmark);
    },
  );
}

/**
 * The url params that a home bookmark opens, from the params or the legacy proto state it was saved with.
 * A param the bookmark does not hold is filled from the explore's defaults when the dashboard loads, other than the filter.
 */
export function getUrlParamsFromHomeBookmark(bookmark: V1Bookmark) {
  const urlParams = bookmark.data
    ? convertProtoToUrlParams(bookmark.data)
    : normalizeExploreUrlParams(new URLSearchParams(bookmark.urlSearch ?? ""));
  // A bookmark keeps the filter it was saved with, and one saved with no filter applied can hold no filter param.
  // It must still open unfiltered, so an empty param keeps the explore's default filter from filling it.
  if (!urlParams.has(ExploreStateURLParams.Filters)) {
    urlParams.set(ExploreStateURLParams.Filters, "");
  }
  return urlParams;
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
