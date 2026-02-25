import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import type { MockUser } from "@rilldata/web-common/features/dashboards/granular-access-policies/useMockUsers";
import { runtimeServiceIssueDevJWT } from "@rilldata/web-common/runtime-client";
import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";

export async function updateDevJWT(
  queryClient: QueryClient,
  instanceId: string,
  mockUser: MockUser | null,
  runtimeClient?: RuntimeClient,
) {
  selectedMockUserStore.set(mockUser);

  if (mockUser === null) {
    selectedMockUserJWT.set(null);
    runtimeClient?.updateJwt(undefined, "user");
    // BRIDGE (temporary): keep global store in sync for unmigrated consumers
    runtime.update((runtimeState) => {
      runtimeState.jwt = undefined;
      return runtimeState;
    });
  } else {
    try {
      const { name, email, groups, admin, ...customAttributes } = mockUser;

      const { jwt } = await runtimeServiceIssueDevJWT({
        email,
        name: name || "Mock User",
        admin: !!admin,
        groups: groups || [],
        attributes: customAttributes,
      });

      if (!jwt) throw new Error("No JWT returned");

      selectedMockUserJWT.set(jwt);
      runtimeClient?.updateJwt(jwt, "mock");
      // BRIDGE (temporary): keep global store in sync for unmigrated consumers
      runtime.update((runtimeState) => {
        runtimeState.jwt = {
          token: jwt,
          receivedAt: Date.now(),
          authContext: "mock",
        };
        return runtimeState;
      });
    } catch {
      // no-op
    }
  }

  return invalidateAllMetricsViews(queryClient, instanceId);
}
