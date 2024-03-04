import { page } from "$app/stores";
import {
  createAdminServiceCreateBookmark,
  createAdminServiceUpdateBookmark,
} from "@rilldata/web-admin/client";
import {
  getBookmarks,
  useProjectId,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useQueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function createHomeBookmarkModifier() {
  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();
  const projectIdRes = useProjectId(
    get(page).params.organization,
    get(page).params.project,
  );
  const bookmarksRes = getBookmarks(
    useQueryClient(),
    get(page).params.organization,
    get(page).params.project,
    get(page).params.dashboard,
  );

  return (data: string) => {
    const bookmarks = get(bookmarksRes);
    const projectId = get(projectIdRes);
    if (bookmarks.isFetching || projectId.isFetching) {
      return;
    }

    if (bookmarks.data?.home) {
      return get(bookmarkUpdater).mutateAsync({
        data: {
          bookmarkId: bookmarks.data.home.id,
          displayName: "Home",
          description: "Main view For this dashboard",
          shared: true,
          default: true,
          data,
        },
      });
    } else {
      return get(bookmarkCreator).mutateAsync({
        data: {
          displayName: "Home",
          description: "Main view For this dashboard",
          projectId: projectId.data,
          resourceKind: ResourceKind.MetricsView,
          resourceName: get(page).params.dashboard,
          shared: true,
          default: true,
          data,
        },
      });
    }
  };
}
