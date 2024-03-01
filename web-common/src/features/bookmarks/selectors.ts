import {
  createAdminServiceGetProject,
  createAdminServiceListBookmarks,
  type V1Bookmark,
} from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { QueryClient } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

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

export type Bookmarks = {
  home: V1Bookmark | undefined;
  personal: V1Bookmark[];
  shared: V1Bookmark[];
};
export function getBookmarks(
  queryClient: QueryClient,
  orgName: string,
  projectName: string,
  metricsViewName: string,
): CreateQueryResult<Bookmarks> {
  return derived([useProjectId(orgName, projectName)], ([projectId], set) =>
    createAdminServiceListBookmarks(
      {
        projectId: projectId.data,
        resourceKind: ResourceKind.MetricsView,
        resourceName: metricsViewName,
      },
      {
        query: {
          enabled: !!projectId?.data && !!metricsViewName,
          select: (resp) => {
            const bookmarks: Bookmarks = {
              home: undefined,
              personal: [],
              shared: [],
            };
            resp.bookmarks?.forEach((bookmark) => {
              if (bookmark.default) {
                bookmarks.home = bookmark;
              } else if (bookmark.shared) {
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
