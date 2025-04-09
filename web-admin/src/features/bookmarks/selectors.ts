import { page } from "$app/stores";
import {
  adminServiceListBookmarks,
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  getAdminServiceListBookmarksQueryKey,
  type V1Bookmark,
  type V1ListBookmarksResponse,
} from "@rilldata/web-admin/client";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { getDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/stores/get-default-explore-url-params";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  getTimeControlState,
  timeControlStateSelector,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlSearch } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-search";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewSchema,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1StructType,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { QueryClient } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";

export type BookmarkEntry = {
  resource: V1Bookmark;
  filtersOnly: boolean;
  absoluteTimeRange: boolean;
  url: string;
};

export type Bookmarks = {
  home: BookmarkEntry | undefined;
  personal: BookmarkEntry[];
  shared: BookmarkEntry[];
};

export async function fetchBookmarks(projectId: string, exploreName: string) {
  const params = {
    projectId,
    resourceKind: ResourceKind.Explore,
    resourceName: exploreName,
  };
  const bookmarksResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListBookmarksQueryKey(params),
    queryFn: ({ signal }) => adminServiceListBookmarks(params, signal),
  });
  return bookmarksResp.bookmarks ?? [];
}

export function getBookmarks(projectId: string, exploreName: string) {
  return derived(createAdminServiceGetCurrentUser(), (userResp, set) =>
    createAdminServiceListBookmarks(
      {
        projectId,
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
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

export function categorizeBookmarks(
  bookmarkResp: V1Bookmark[],
  metricsSpec: V1MetricsViewSpec | undefined,
  exploreSpec: V1ExploreSpec | undefined,
  schema: V1StructType | undefined,
  exploreState: MetricsExplorerEntity,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const bookmarks: Bookmarks = {
    home: undefined,
    personal: [],
    shared: [],
  };
  if (!exploreState) return bookmarks;

  const defaultExploreUrlParams = getDefaultExploreUrlParams(
    metricsSpec,
    exploreSpec,
    timeRangeSummary,
  );

  bookmarkResp?.forEach((bookmarkResource) => {
    const bookmark = parseBookmark(
      bookmarkResource,
      metricsSpec ?? {},
      exploreSpec ?? {},
      schema ?? {},
      exploreState,
      defaultExploreUrlParams,
      timeRangeSummary,
    );
    if (isHomeBookmark(bookmarkResource)) {
      bookmarks.home = bookmark;
    } else if (bookmarkResource.shared) {
      bookmarks.shared.push(bookmark);
    } else {
      bookmarks.personal.push(bookmark);
    }
  });

  return bookmarks;
}

export function getHomeBookmarkExploreState(
  projectId: string,
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
): CompoundQueryResult<Partial<MetricsExplorerEntity> | null> {
  return getCompoundQuery(
    [
      getBookmarks(projectId, exploreName),
      useExploreValidSpec(instanceId, exploreName),
      createQueryServiceMetricsViewSchema(instanceId, metricsViewName),
    ],
    ([bookmarksResp, exploreSpecResp, schemaResp]) => {
      const homeBookmark = bookmarksResp?.bookmarks?.find(isHomeBookmark);
      if (!homeBookmark) return null;

      const exploreSpec = exploreSpecResp?.explore ?? {};
      const metricsViewSpec = exploreSpecResp?.metricsView ?? {};

      const exploreStateFromHomeBookmark = getDashboardStateFromUrl(
        homeBookmark?.data ?? "",
        metricsViewSpec,
        exploreSpec,
        schemaResp?.schema ?? {},
      );
      return exploreStateFromHomeBookmark;
    },
  );
}

export function searchBookmarks(
  bookmarks: Bookmarks | undefined,
  searchText: string,
): Bookmarks | undefined {
  if (!searchText || !bookmarks) return bookmarks;
  searchText = searchText.toLowerCase();
  const matchName = (bookmark: BookmarkEntry | undefined) =>
    bookmark?.resource.displayName &&
    bookmark.resource.displayName.toLowerCase().includes(searchText);
  return {
    home: matchName(bookmarks.home) ? bookmarks.home : undefined,
    personal: bookmarks?.personal.filter(matchName) ?? [],
    shared: bookmarks?.shared.filter(matchName) ?? [],
  };
}

export function getPrettySelectedTimeRange(
  queryClient: QueryClient,
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
): Readable<string> {
  return derived(
    [
      useExploreValidSpec(instanceId, exploreName),
      useMetricsViewTimeRange(instanceId, metricsViewName, {}, queryClient),
      useExploreState(metricsViewName),
    ],
    ([validSpec, timeRangeSummary, metricsExplorerEntity]) => {
      const timeRangeState = timeControlStateSelector([
        validSpec.data?.metricsView ?? {},
        validSpec.data?.explore ?? {},
        timeRangeSummary,
        metricsExplorerEntity,
      ]);
      if (!timeRangeState.ready) return "";
      return prettyFormatTimeRange(
        timeRangeState.selectedTimeRange?.start,
        timeRangeState.selectedTimeRange?.end,
        timeRangeState.selectedTimeRange?.name,
        metricsExplorerEntity?.selectedTimezone,
      );
    },
  );
}

function parseBookmark(
  bookmarkResource: V1Bookmark,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  schema: V1StructType,
  exploreState: MetricsExplorerEntity,
  defaultExploreUrlParams: URLSearchParams,
  timeRangeSummary: V1TimeRangeSummary | undefined,
): BookmarkEntry {
  const exploreStateFromBookmark = getDashboardStateFromUrl(
    bookmarkResource.data ?? "",
    metricsViewSpec,
    exploreSpec,
    schema,
  );

  const finalExploreState = {
    ...(exploreState ?? {}),
    ...exploreStateFromBookmark,
  } as MetricsExplorerEntity;

  const url = new URL(get(page).url);

  const searchParams = convertPartialExploreStateToUrlSearch(
    finalExploreState,
    exploreSpec,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
      finalExploreState,
    ),
    defaultExploreUrlParams,
    url,
  );

  url.search = searchParams.toString();

  return {
    resource: bookmarkResource,
    absoluteTimeRange:
      exploreStateFromBookmark.selectedTimeRange?.name ===
      TimeRangePreset.CUSTOM,
    filtersOnly: isFilterOnlyBookmark(
      exploreStateFromBookmark,
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
      defaultExploreUrlParams,
      url,
    ),
    url: url.toString(),
  };
}

function isFilterOnlyBookmark(
  bookmarkState: Partial<MetricsExplorerEntity>,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  defaultExploreUrlParams: URLSearchParams,
  url: URL,
): boolean {
  // Get the bookmark's search params
  const searchParams = convertPartialExploreStateToUrlSearch(
    bookmarkState as MetricsExplorerEntity,
    exploreSpec,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
      bookmarkState as MetricsExplorerEntity,
    ),
    defaultExploreUrlParams,
    url,
  );

  // These are the only parameters that are stored in a filter-only bookmark
  const allowedFilterParams = new Set([
    ExploreStateURLParams.Filters,
    ExploreStateURLParams.TimeRange,
    ExploreStateURLParams.TimeGrain,
  ]) as Set<string>;

  // Check if all the bookmark's search params are in the allowed list
  const urlParams = Array.from(searchParams.keys());
  return urlParams.every((param) => allowedFilterParams.has(param));
}
