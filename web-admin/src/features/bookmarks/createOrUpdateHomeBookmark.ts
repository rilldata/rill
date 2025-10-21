import {
  createAdminServiceCreateBookmark,
  createAdminServiceUpdateBookmark,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { isHomeBookmark } from "@rilldata/web-admin/features/bookmarks/selectors";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { get } from "svelte/store";

// TODO: move this to a compound mutation similar to ProjectGithubConnectionUpdater
export function createHomeBookmarkModifier(
  resourceKind: ResourceKind,
  resourceName: string,
) {
  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  return (urlSearch: string, projectId: string, bookmarks: V1Bookmark[]) => {
    const homeBookmark = bookmarks.find(isHomeBookmark);

    if (homeBookmark) {
      return get(bookmarkUpdater).mutateAsync({
        data: {
          bookmarkId: homeBookmark.id,
          displayName: "Go to Home",
          description: "",
          shared: true,
          default: true,
          urlSearch,
        },
      });
    } else {
      return get(bookmarkCreator).mutateAsync({
        data: {
          displayName: "Go to Home",
          description: "",
          projectId,
          resourceKind,
          resourceName,
          shared: true,
          default: true,
          urlSearch,
        },
      });
    }
  };
}
