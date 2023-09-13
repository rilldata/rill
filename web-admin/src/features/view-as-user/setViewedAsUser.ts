import {
  adminServiceGetDeploymentCredentials,
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
  user: V1User
) {
  viewAsUserStore.set(user);

  let jwt: string = null;
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

  runtime.update((runtimeState) => {
    runtimeState.jwt = jwt;
    return runtimeState;
  });

  invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}
