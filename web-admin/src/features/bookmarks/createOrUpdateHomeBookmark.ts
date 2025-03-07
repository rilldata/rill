import {
  createAdminServiceCreateBookmark,
  createAdminServiceUpdateBookmark,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { isHomeBookmark } from "@rilldata/web-admin/features/bookmarks/selectors";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { get } from "svelte/store";

// TODO: move this to a compound mutation similar to ProjectGithubConnectionUpdater
export function createHomeBookmarkModifier(exploreName: string) {
  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  return (data: string, projectId: string, bookmarks: V1Bookmark[]) => {
    const homeBookmark = bookmarks.find(isHomeBookmark);

    if (homeBookmark) {
      return get(bookmarkUpdater).mutateAsync({
        data: {
          bookmarkId: homeBookmark.id,
          displayName: "Home",
          description: "Main view for this dashboard",
          shared: true,
          default: true,
          data,
        },
      });
    } else {
      return get(bookmarkCreator).mutateAsync({
        data: {
          displayName: "Home",
          description: "Main view for this dashboard",
          projectId,
          resourceKind: ResourceKind.Explore,
          resourceName: exploreName,
          shared: true,
          default: true,
          data,
        },
      });
    }
  };
}
