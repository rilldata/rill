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
    queryClient.refetchQueries({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),

      // avoid refetching createAdminServiceGetProjectWithBearerToken
      exact: true,
    }),
    queryClient.refetchQueries({
      queryKey: getAdminServiceGetGithubUserStatusQueryKey(),
    }),
    invalidateRuntimeQueries(queryClient, instanceId),
  ]);
}
