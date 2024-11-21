import {
  adminServiceListBookmarks,
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  getAdminServiceListBookmarksQueryKey,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { useExploreStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import {
  createQueryServiceMetricsViewSchema,
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1StructType,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type BookmarkEntry = {
  resource: V1Bookmark;
  metricsEntity: Partial<MetricsExplorerEntity>;
  filtersOnly: boolean;
  absoluteTimeRange: boolean;
  url: string;
};

export type Bookmarks = {
  home: BookmarkEntry | undefined;
  personal: BookmarkEntry[];
  shared: BookmarkEntry[];
};
export function getBookmarks(
  queryClient: QueryClient,
  instanceId: string,
  orgName: string,
  projectName: string,
  metricsViewName: string,
  exploreName: string,
): CreateQueryResult<Bookmarks> {
  return derived(
    [
      useProjectId(orgName, projectName),
      useExploreValidSpec(instanceId, exploreName),
      createQueryServiceMetricsViewSchema(instanceId, metricsViewName),
      createAdminServiceGetCurrentUser(),
    ],
    ([projectId, validSpec, schemaResp, userResp], set) =>
      createAdminServiceListBookmarks(
        {
          projectId: projectId.data,
          resourceKind: ResourceKind.Explore,
          resourceName: exploreName,
        },
        {
          query: {
            enabled:
              !!projectId?.data &&
              !!metricsViewName &&
              !!exploreName &&
              !validSpec.isFetching &&
              !schemaResp.isFetching &&
              userResp.isSuccess &&
              !!userResp.data.user,
            select: (resp) =>
              categorizeBookmarks(
                resp.bookmarks ?? [],
                validSpec.data?.metricsView,
                validSpec.data?.explore,
                schemaResp.data?.schema,
              ),
            queryClient,
          },
        },
      ).subscribe(set),
  );
}

export async function fetchBookmarks(
  projectId: string,
  exploreName: string,
  metricsSpec: V1MetricsViewSpec | undefined,
  exploreSpec: V1ExploreSpec | undefined,
) {
  const params = {
    projectId,
    resourceKind: ResourceKind.Explore,
    resourceName: exploreName,
  };
  const bookmarksResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListBookmarksQueryKey(params),
    queryFn: ({ signal }) => adminServiceListBookmarks(params, signal),
  });
  return categorizeBookmarks(
    bookmarksResp.bookmarks ?? [],
    metricsSpec,
    exploreSpec,
    undefined, // TODO
  );
}

function categorizeBookmarks(
  bookmarkResp: V1Bookmark[],
  metricsSpec: V1MetricsViewSpec | undefined,
  exploreSpec: V1ExploreSpec | undefined,
  schema: V1StructType | undefined,
) {
  const bookmarks: Bookmarks = {
    home: undefined,
    personal: [],
    shared: [],
  };
  bookmarkResp?.forEach((bookmarkResource) => {
    const bookmark = parseBookmarkEntry(
      bookmarkResource,
      metricsSpec ?? {},
      exploreSpec ?? {},
      schema ?? {},
    );
    if (bookmarkResource.default) {
      bookmarks.home = bookmark;
    } else if (bookmarkResource.shared) {
      bookmarks.shared.push(bookmark);
    } else {
      bookmarks.personal.push(bookmark);
    }
  });
  return bookmarks;
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

export function getHomeBookmarkData(
  queryClient: QueryClient,
  instanceId: string,
  orgName: string,
  projectName: string,
  metricsViewName: string,
  exploreName: string,
): CompoundQueryResult<string> {
  return derived(
    getBookmarks(
      queryClient,
      instanceId,
      orgName,
      projectName,
      metricsViewName,
      exploreName,
    ),
    (bookmarks) => {
      if (bookmarks.isFetching || !bookmarks.data) {
        return {
          isFetching: true,
          error: "",
        };
      } else if (bookmarks.isError) {
        return {
          isFetching: false,
          error: bookmarks.error,
        };
      }
      return {
        isFetching: false,
        error: "",
        data: bookmarks.data?.home?.resource?.data,
      };
    },
  );
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
      useMetricsViewTimeRange(instanceId, metricsViewName, {
        query: { queryClient },
      }),
      useExploreStore(metricsViewName),
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

export function getFilledInBookmarks(
  bookmarks: Bookmarks | undefined,
  baseUrl: string,
  dashboard: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  basePreset: V1ExplorePreset,
): Bookmarks | undefined {
  if (!bookmarks) return undefined;

  if (!baseUrl.startsWith("http")) {
    // handle case where only path is provided
    baseUrl = "http://localhost" + baseUrl;
  }

  return {
    home: bookmarks.home
      ? getFilledInBookmark(
          bookmarks.home,
          baseUrl,
          dashboard,
          exploreSpec,
          basePreset,
        )
      : undefined,
    personal: bookmarks.personal.map((b) =>
      getFilledInBookmark(b, baseUrl, dashboard, exploreSpec, basePreset),
    ),
    shared: bookmarks.shared.map((b) =>
      getFilledInBookmark(b, baseUrl, dashboard, exploreSpec, basePreset),
    ),
  };
}

function parseBookmarkEntry(
  bookmarkResource: V1Bookmark,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  schema: V1StructType,
): BookmarkEntry {
  const metricsEntity = getDashboardStateFromUrl(
    bookmarkResource.data ?? "",
    metricsViewSpec,
    exploreSpec,
    schema,
  );
  return {
    resource: bookmarkResource,
    metricsEntity,
    absoluteTimeRange:
      metricsEntity.selectedTimeRange?.name === TimeRangePreset.CUSTOM,
    filtersOnly: !metricsEntity.pivot,
    url: "", // will be filled in along with existing dashboard
  };
}

function getFilledInBookmark(
  bookmark: BookmarkEntry,
  baseUrl: string,
  dashboard: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  basePreset: V1ExplorePreset,
) {
  const url = new URL(baseUrl);
  const searchParamsFromBookmark = convertMetricsEntityToURLSearchParams(
    { ...dashboard, ...bookmark.metricsEntity },
    exploreSpec,
    basePreset,
  );
  mergeSearchParams(url.searchParams, searchParamsFromBookmark);
  return {
    ...bookmark,
    url: `${url.pathname}${url.search}`,
  };
}
