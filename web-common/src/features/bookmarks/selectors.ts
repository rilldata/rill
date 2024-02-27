import {
  createAdminServiceGetProject,
  createAdminServiceListBookmarks,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import type { QueryClient } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function useProjectId(orgName: string, projectName: string) {
  return createAdminServiceGetProject(orgName, projectName, {
    query: {
      enabled: !!orgName && !!projectName,
      select: (resp) => resp.project?.id,
    },
  });
}

export type Bookmarks = {
  home: V1Bookmark | undefined;
  own: V1Bookmark[];
  global: V1Bookmark[];
};
export function getBookmarks(
  queryClient: QueryClient,
  orgName: string,
  projectName: string,
  dashboardName: string,
): CreateQueryResult<Bookmarks> {
  return derived([useProjectId(orgName, projectName)], ([projectId], set) =>
    createAdminServiceListBookmarks(
      {
        projectId: projectId.data,
        dashboardName,
      },
      {
        query: {
          enabled: !!projectId?.data && !!dashboardName,
          select: (resp) => {
            const bookmarks: Bookmarks = {
              home: undefined,
              own: resp.bookmarks?.filter((b) => !b.isGlobal) ?? [],
              global: resp.bookmarks?.filter((b) => b.isGlobal) ?? [],
            };
            const homeIndex = bookmarks.global.findIndex(
              (b) => b.displayName === "Home",
            );
            if (homeIndex >= 0) {
              bookmarks.home = bookmarks.global.splice(homeIndex, 1)[0];
            }

            return bookmarks;
          },
          queryClient,
        },
      },
    ).subscribe(set),
  );
}
