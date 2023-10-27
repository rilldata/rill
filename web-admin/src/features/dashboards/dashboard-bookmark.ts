import {
  createAdminServiceCreateBookmark,
  createAdminServiceListBookmarks,
  createAdminServiceUpdateBookmark,
  getAdminServiceListBookmarksQueryOptions,
} from "@rilldata/web-admin/client";
import { Debounce } from "@rilldata/web-common/features/models/utils/Debounce";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export const DashboardSaveStateBookmark = "__save_state__";

// TODO: do we need a list query with dashboard as argument?
export function useDashboardSaveState(
  projectId: string,
  dashboardName: string
) {
  return createAdminServiceListBookmarks(
    {
      projectId,
    },
    {
      query: {
        select: (data) =>
          data.bookmarks.find(
            (b) =>
              b.dashboardName === dashboardName &&
              b.displayName === DashboardSaveStateBookmark
          ),
        enabled: !!dashboardName && !!projectId,
        keepPreviousData: true,
      },
    }
  );
}

export function createDashboardSaveStateInit(
  queryClient: QueryClient,
  projectId: string,
  dashboardName: string
) {
  const createMutation = createAdminServiceCreateBookmark();
  let created = false;

  return async (proto: string) => {
    if (created) return;
    await get(createMutation).mutateAsync({
      data: {
        projectId,
        dashboardName,
        data: proto,
        displayName: DashboardSaveStateBookmark,
      },
    });
    created = true;
    return queryClient.resetQueries(
      getAdminServiceListBookmarksQueryOptions({
        projectId,
      })
    );
  };
}

export function createDashboardSaveStateMutation(
  queryClient: QueryClient,
  projectId: string,
  dashboardName: string
) {
  const existingBookmark = useDashboardSaveState(projectId, dashboardName);
  const updateMutation = createAdminServiceUpdateBookmark();
  const debouncer = new Debounce();

  return async (proto: string) => {
    debouncer.debounce(
      dashboardName,
      async () => {
        await waitUntil(() => {
          const bookmark = get(existingBookmark);
          return !bookmark.isFetching;
        });
        const bookmark = get(existingBookmark);
        if (bookmark.data.data === proto) return;
        await get(updateMutation).mutateAsync({
          bookmarkId: bookmark.data.id,
          data: {
            data: proto,
            displayName: DashboardSaveStateBookmark,
          },
        });
        return queryClient.resetQueries(
          getAdminServiceListBookmarksQueryOptions({
            projectId,
          })
        );
      },
      500
    );
  };
}
