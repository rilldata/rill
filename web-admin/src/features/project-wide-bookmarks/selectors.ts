import {
  getProjectIdQueryOptions,
  type OrgAndProjectNameStore,
} from "@rilldata/web-admin/features/projects/selectors.ts";
import {
  createAdminServiceGetCurrentUser,
  getAdminServiceListBookmarksInfiniteQueryOptions,
} from "@rilldata/web-admin/client";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

const BookmarksPageSize = 1000;

export function getBookmarksInfiniteQueryOptions(
  orgAndProjectNameStore: OrgAndProjectNameStore,
) {
  const projectIdQuery = createQuery(
    getProjectIdQueryOptions(orgAndProjectNameStore),
  );

  return derived(
    [createAdminServiceGetCurrentUser(), projectIdQuery],
    ([userResp, projectIdQueryResp]) => {
      const hasUser = userResp.data?.user;
      const projectId = projectIdQueryResp.data;

      return getAdminServiceListBookmarksInfiniteQueryOptions(
        {
          projectId,
          pageSize: BookmarksPageSize,
        },
        {
          query: {
            enabled: Boolean(hasUser && !!projectId),
            getNextPageParam: (lastPage) => {
              if (lastPage.nextPageToken !== "") {
                return lastPage.nextPageToken;
              }
              return undefined;
            },
          },
        },
      );
    },
  );
}
