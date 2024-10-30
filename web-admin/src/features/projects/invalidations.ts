import {
  getAdminServiceGetGithubUserStatusQueryKey,
  getAdminServiceGetProjectQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";

export function invalidateProjectQueries(
  instanceId: string,
  organization: string,
  project: string,
) {
  return Promise.all([
    queryClient.refetchQueries(
      getAdminServiceGetProjectQueryKey(organization, project),
      {
        // avoid refetching createAdminServiceGetProjectWithBearerToken
        exact: true,
      },
    ),
    queryClient.refetchQueries(getAdminServiceGetGithubUserStatusQueryKey()),
    invalidateRuntimeQueries(queryClient, instanceId),
  ]);
}
