import {
  createAdminServiceGetCurrentUser,
  createAdminServiceListBookmarks,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import {
  useMetricsView,
  useMetricsViewTimeRange,
} from "@rilldata/web-common/features/dashboards/selectors";
import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewSchema,
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
): CreateQueryResult<Bookmarks> {
  return derived(
    [
      useProjectId(orgName, projectName),
      useMetricsView(instanceId, metricsViewName),
      createQueryServiceMetricsViewSchema(instanceId, metricsViewName),
      createAdminServiceGetCurrentUser(),
    ],
    ([projectId, metricsViewResp, schemaResp, userResp], set) =>
      createAdminServiceListBookmarks(
        {
          projectId: projectId.data,
          resourceKind: ResourceKind.MetricsView,
          resourceName: metricsViewName,
        },
        {
          query: {
            enabled:
              !!projectId?.data &&
              !!metricsViewName &&
              !metricsViewResp.isFetching &&
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
                  metricsViewResp.data as V1MetricsViewSpec,
                  schemaResp.data?.schema as V1StructType,
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
): CompoundQueryResult<string> {
  return derived(
    getBookmarks(
      queryClient,
      instanceId,
      orgName,
      projectName,
      metricsViewName,
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
): Readable<string> {
  return derived(
    [
      useMetricsView(instanceId, metricsViewName),
      useMetricsViewTimeRange(instanceId, metricsViewName, {
        query: { queryClient },
      }),
      useDashboardStore(metricsViewName),
    ],
    ([metricViewSpec, timeRangeSummary, metricsExplorerEntity]) => {
      const timeRangeState = timeControlStateSelector([
        metricViewSpec,
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

function parseBookmarkEntry(
  bookmarkResource: V1Bookmark,
  metricsViewSpec: V1MetricsViewSpec,
  schema: V1StructType,
): BookmarkEntry {
  const metricsEntity = getDashboardStateFromUrl(
    bookmarkResource.data ?? "",
    metricsViewSpec,
    schema,
  );
  return {
    resource: bookmarkResource,
    absoluteTimeRange:
      metricsEntity.selectedTimeRange?.name === TimeRangePreset.CUSTOM,
    filtersOnly: !metricsEntity.pivot,
  };
}
