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

  if (mockUser === null) {
    selectedMockUserJWT.set(null);
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

  return invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}
