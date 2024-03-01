import {
  createAdminServiceGetProject,
  createAdminServiceListBookmarks,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewSchema,
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewSpec,
  type V1StructType,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function useProjectId(orgName: string, projectName: string) {
  return createAdminServiceGetProject(
    orgName,
    projectName,
    {},
    {
      query: {
        enabled: !!orgName && !!projectName,
        select: (resp) => resp.project?.id,
      },
    },
  );
}

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
    ],
    ([projectId, metricsViewResp, schemaResp], set) =>
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
              !schemaResp.isFetching,
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

export function getPrettySelectedTimeRange(
  queryClient: QueryClient,
  instanceId: string,
  metricsViewName: string,
): Readable<string> {
  return derived(
    [
      useMetricsView(instanceId, metricsViewName),
      derived(
        [runtime, useMetricsView(instanceId, metricsViewName)],
        ([runtime, metricsView], set) =>
          createQueryServiceMetricsViewTimeRange(
            runtime.instanceId,
            metricsViewName,
            {},
            {
              query: {
                queryClient: queryClient,
                enabled: !!metricsView.data?.timeDimension,
              },
            },
          ).subscribe(set),
      ),
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
        metricsExplorerEntity.selectedTimezone,
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
    filtersOnly: !metricsEntity.selectedTimeRange && !metricsEntity.pivot,
  };
}
