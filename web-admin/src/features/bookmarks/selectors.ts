import {
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { useExploreStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { getTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/time-ranges";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewSchema,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1StructType,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type BookmarkEntry = {
  resource: V1Bookmark;
  filtersOnly: boolean;
  absoluteTimeRange: boolean;
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
            select: (resp) => {
              const bookmarks: Bookmarks = {
                home: undefined,
                personal: [],
                shared: [],
              };
              resp.bookmarks?.forEach((bookmarkResource) => {
                const bookmark = parseBookmarkEntry(
                  bookmarkResource,
                  validSpec.data?.metricsView ?? {},
                  validSpec.data?.explore ?? {},
                  schemaResp.data?.schema ?? {},
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
            },
            queryClient,
          },
        },
      ).subscribe(set),
  );
}

export function searchBookmarks(
  bookmarks: Bookmarks,
  searchText: string,
): Bookmarks {
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
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
): Readable<string> {
  return derived(
    [
      useExploreValidSpec(instanceId, exploreName),
      getTimeRanges(exploreName),
      useExploreStore(metricsViewName),
    ],
    ([validSpec, timeRanges, metricsExplorerEntity]) => {
      const timeRangeState = timeControlStateSelector([
        validSpec.data?.metricsView ?? {},
        validSpec.data?.explore ?? {},
        timeRanges,
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
    absoluteTimeRange:
      metricsEntity.selectedTimeRange?.name === TimeRangePreset.CUSTOM,
    filtersOnly: !metricsEntity.pivot,
  };
}
