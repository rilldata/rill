import {
  getAdminServiceListBookmarksQueryKey,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

export const BookmarksPageSize = 1000;

export type Bookmarks = {
  home: V1Bookmark | undefined;
  personal: V1Bookmark[];
  shared: V1Bookmark[];
};

export function categorizeBookmarks(bookmarkResources: V1Bookmark[]) {
  const bookmarks: Bookmarks = {
    home: undefined,
    personal: [],
    shared: [],
  };

  bookmarkResources.forEach((bookmark) => {
    if (bookmark.default) {
      bookmarks.home = bookmark;
    } else if (bookmark.shared) {
      bookmarks.shared.push(bookmark);
    } else {
      bookmarks.personal.push(bookmark);
    }
  });

  return bookmarks;
}

export function invalidateBookmarksQuery(projectId: string) {
  return queryClient.refetchQueries({
    queryKey: getAdminServiceListBookmarksQueryKey({
      projectId,
      pageSize: BookmarksPageSize,
    }),
  });
}
