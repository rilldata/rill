import {
  adminServiceGetDeploymentCredentials,
  adminServiceGetProject,
  getAdminServiceGetProjectQueryKey,
  V1GetProjectResponse,
  type V1User,
} from "@rilldata/web-admin/client";
import { viewAsUserStore } from "@rilldata/web-admin/components/authentication/viewAsUserStore";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export async function updateMimickedJWT(
  queryClient: QueryClient,
  organization: string,
  project: string,
  user: V1User | null
) {
  viewAsUserStore.set(user);
  let jwt: string = null;

  if (user !== null) {
    try {
      const jwtResp = await adminServiceGetDeploymentCredentials(
        organization,
        project,
        {
          userId: user.id,
        }
      );
      jwt = jwtResp.jwt;
    } catch (e) {
      // no-op
    }
  } else {
    // Get the admin's JWT from `GetProject` call
    const projResp = await queryClient.fetchQuery<V1GetProjectResponse>({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      queryFn: () => adminServiceGetProject(organization, project),
    });
    jwt = projResp.jwt;
  }

  runtime.update((runtimeState) => {
    runtimeState.jwt = jwt;
    return runtimeState;
  });
  return invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}
