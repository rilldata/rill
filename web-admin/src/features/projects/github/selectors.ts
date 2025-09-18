import {
  createAdminServiceGetGithubUserStatus,
  getAdminServiceListGithubUserReposQueryOptions,
  V1GithubPermission,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getGithubUserOrgs() {
  return createAdminServiceGetGithubUserStatus({
    query: {
      select: (data) => {
        const orgs = Object.entries(
          data.organizationInstallationPermissions ?? {},
        )
          // Filter only org we have write permission to
          .filter(
            ([, permission]) =>
              permission === V1GithubPermission.GITHUB_PERMISSION_WRITE,
          )
          .map(([org]) => org);
        const hasWritePermissionToUserOrg =
          data.userInstallationPermission ===
          V1GithubPermission.GITHUB_PERMISSION_WRITE;
        if (hasWritePermissionToUserOrg && data.account) {
          orgs.push(data.account);
        }

        return orgs.map((o) => ({ value: o, label: o }));
      },
    },
  });
}

export function getGithubUserRepos(enabled: boolean) {
  const userOrgsOpts = derived(
    createAdminServiceGetGithubUserStatus(undefined, queryClient),
    (userStatus) =>
      getAdminServiceListGithubUserReposQueryOptions({
        query: {
          enabled:
            !!userStatus.data?.hasAccess && !userStatus.isFetching && enabled,
          select: (data) => ({
            rawRepos: data.repos ?? [],
            repoOptions:
              data.repos?.map((repo) => ({
                value: repo.remote,
                label: `${repo.owner}/${repo.name}`,
              })) ?? [],
          }),
        },
      }),
  );

  return createQuery(userOrgsOpts);
}
