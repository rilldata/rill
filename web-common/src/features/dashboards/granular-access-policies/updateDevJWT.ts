import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import { runtimeServiceIssueDevJWT } from "@rilldata/web-common/runtime-client";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export async function updateDevJWT(
  queryClient: QueryClient,
  mockUser: MockUser | null,
) {
  selectedMockUserStore.set(mockUser);
  let jwt: string = null;

  if (mockUser !== null) {
    try {
      const jwtResp = await runtimeServiceIssueDevJWT({
        name: mockUser?.name ? mockUser.name : "Mock User",
        email: mockUser?.email,
        groups: mockUser?.groups ? mockUser.groups : [],
        admin: !!mockUser?.admin,
      });
      jwt = jwtResp.jwt;
    } catch (e) {
      // no-op
    }
  }

  selectedMockUserJWT.set(jwt);
  runtime.update((runtimeState) => {
    runtimeState.jwt = {
      token: jwt,
      receivedAt: Date.now(),
    };
    return runtimeState;
  });
  return invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}
