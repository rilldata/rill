import {
  adminServiceGetDeploymentCredentials,
  getAdminServiceGetDeploymentCredentialsQueryKey,
  V1GetDeploymentCredentialsResponse,
  type V1User,
} from "@rilldata/web-admin/client";
import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export async function setViewedAsUser(
  queryClient: QueryClient,
  organization: string,
  project: string,
  user: V1User,
) {
  viewAsUserStore.set(user);

  const jwtResp =
    await queryClient.fetchQuery<V1GetDeploymentCredentialsResponse>({
      queryKey: getAdminServiceGetDeploymentCredentialsQueryKey(
        organization,
        project,
        {
          userId: user.id,
        },
      ),
      queryFn: () =>
        adminServiceGetDeploymentCredentials(organization, project, {
          userId: user.id,
        }),
    });
  const jwt = jwtResp.accessToken;

  runtime.update((runtimeState) => {
    runtimeState.jwt = {
      token: jwt,
      receivedAt: Date.now(),
    };
    return runtimeState;
  });

  await invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}
